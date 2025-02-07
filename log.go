package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

var loggerSingleton *logger

// Level defined the type for a log level.
type Level int

const (
	// LogLevelDebug debug log level.
	LogLevelDebug Level = 0

	// LogLevelInfo info log level.
	LogLevelInfo Level = 1

	// LogLevelWarn warn log level.
	LogLevelWarn Level = 2

	// LogLevelError error log level.
	LogLevelError Level = 3

	// LogLevelFatal fatal log level.
	LogLevelFatal Level = 4
)

// String returns the actual currency string to be used in the wallet
func (t Level) String() string {
	return [...]string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}[t]
}

type logger struct {
	token         string
	Level         Level
	url           string
	bulk          bool
	bufferSize    int
	flushInterval time.Duration
	buffer        []*logMessage
	sync.Mutex
	tags      []string
	debugMode bool
}

type logMessage struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data"`
}

// SetupLogger creates a new loggly logger.
func SetupLogger(token string, level Level, tags []string, bulk bool, debugMode bool) {
	if loggerSingleton != nil {
		return
	}

	// Setup logger with options.
	loggerSingleton = &logger{
		token:         token,
		Level:         level,
		url:           "",
		bulk:          bulk,
		bufferSize:    1000,
		flushInterval: 10 * time.Second,
		buffer:        make([]*logMessage, 0, 1000),
		tags:          tags,
		debugMode:     debugMode,
	}

	// If the bulk option is set make sure we set the url to the bulk endpoint.
	if bulk {
		loggerSingleton.url = "https://logs-01.loggly.com/bulk/" + token + "/tag/" + tagList() + "/"

		// Start flush interval
		go start()
	} else {
		loggerSingleton.url = "https://logs-01.loggly.com/inputs/" + token + "/tag/" + tagList() + "/"
	}

}

// Stdln prints the output.
func Stdln(output string) {
	fmt.Println(output)
}

// Stdf prints the formatted output.
func Stdf(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

// Debugln prints the output.
func Debugln(output string) {
	Debugd(output, nil)
}

// Debugd prints output string and data.
func Debugd(output string, data map[string]interface{}) {
	if loggerSingleton == nil || loggerSingleton.Level > LogLevelDebug {
		return
	}

	buildAndShipMessage(output, LogLevelDebug.String(), false, data)
}

// Debugf prints the formatted output.
func Debugf(format string, a ...interface{}) {
	Debugln(fmt.Sprintf(format, a...))
}

// Debugdf prints the formatted output. Special format is used, looking for expressions like @Field in the 'output',
// and replacing with given values. Data then will be logged like {"Field": "value"}. Special format verbs are
// supported with fields, so for example @Number%04d is supported to format integers. Default value is %v.
// Example: format="example @Data message @Used", a=[1234, "text"] will log
// message="example 1234 message text" and data={"Data": 1234, "Used": "text"}
// Returns the formatted message
func Debugdf(format string, values ...interface{}) string {
	message, data := formatDataMessages(format, values...)
	Debugd(message, data)
	return message
}

// Infoln prints the output.
func Infoln(output string) {
	Infod(output, nil)
}

// Infof prints the formatted output.
func Infof(format string, a ...interface{}) {
	Infoln(fmt.Sprintf(format, a...))
}

// Infod prints output string and data.
func Infod(output string, data map[string]interface{}) {
	if loggerSingleton == nil || loggerSingleton.Level > LogLevelInfo {
		return
	}
	buildAndShipMessage(output, LogLevelInfo.String(), false, data)
}

// Infodf prints the formatted output. Special format is used, looking for expressions like @Field in the 'output',
// and replacing with given values. Data then will be logged like {"Field": "value"}. Special format verbs are
// supported with fields, so for example @Number%04d is supported to format integers. Default value is %v.
// Example: format="example @Data message @Used", a=[1234, "text"] will log
// message="example 1234 message text" and data={"Data": 1234, "Used": "text"}
// Returns the formatted message
func Infodf(format string, values ...interface{}) string {
	message, data := formatDataMessages(format, values...)
	Infod(message, data)
	return message
}

// Warnln prints the output.
func Warnln(output string) {
	Warnd(output, nil)
}

// Warnf prints the formatted output.
func Warnf(format string, a ...interface{}) {
	Warnln(fmt.Sprintf(format, a...))
}

// Warnd prints output string and data.
func Warnd(output string, data map[string]interface{}) {
	if loggerSingleton == nil || loggerSingleton.Level > LogLevelWarn {
		return
	}

	buildAndShipMessage(output, LogLevelWarn.String(), false, data)
}

// Warndf prints the formatted output. Special format is used, looking for expressions like @Field in the 'output',
// and replacing with given values. Data then will be logged like {"Field": "value"}. Special format verbs are
// supported with fields, so for example @Number%04d is supported to format integers. Default value is %v.
// Example: format="example @Data message @Used", a=[1234, "text"] will log
// message="example 1234 message text" and data={"Data": 1234, "Used": "text"}
// Returns the formatted message
func Warndf(format string, values ...interface{}) string {
	message, data := formatDataMessages(format, values...)
	Warnd(message, data)
	return message
}

// Errorln prints the output.
func Errorln(output string) {
	Errord(output, nil)
}

// Errorf prints the formatted output.
func Errorf(format string, a ...interface{}) {
	Errorln(fmt.Sprintf(format, a...))
}

// Errord prints output string and data.
func Errord(output string, data map[string]interface{}) {
	if loggerSingleton == nil || loggerSingleton.Level > LogLevelError {
		return
	}

	buildAndShipMessage(output, LogLevelError.String(), false, data)
}

// Errordf prints the formatted output. Special format is used, looking for expressions like @Field in the 'output',
// and replacing with given values. Data then will be logged like {"Field": "value"}. Special format verbs are
// supported with fields, so for example @Number%04d is supported to format integers. Default value is %v.
// Example: format="example @Data message @Used", a=[1234, "text"] will log
// message="example 1234 message text" and data={"Data": 1234, "Used": "text"}
// Returns the formatted message
func Errordf(format string, values ...interface{}) string {
	message, data := formatDataMessages(format, values...)
	Errord(message, data)
	return message
}

// Fatalln prints the output.
func Fatalln(output string) {
	Fatald(output, nil)
}

// Fatalf prints the formatted output.
func Fatalf(format string, a ...interface{}) {
	Fatalln(fmt.Sprintf(format, a...))
}

// Fatald prints output string and data.
func Fatald(output string, data map[string]interface{}) {
	if loggerSingleton == nil || loggerSingleton.Level > LogLevelFatal {
		return
	}

	buildAndShipMessage(output, LogLevelFatal.String(), true, data)
}

// Fataldf prints the formatted output. Special format is used, looking for expressions like @Field in the 'output',
// and replacing with given values. Data then will be logged like {"Field": "value"}. Special format verbs are
// supported with fields, so for example @Number%04d is supported to format integers. Default value is %v.
// Example: format="example @Data message @Used", a=[1234, "text"] will log
// message="example 1234 message text" and data={"Data": 1234, "Used": "text"}
// Returns the formatted message
func Fataldf(format string, values ...interface{}) string {
	message, data := formatDataMessages(format, values...)
	Fatald(message, data)
	return message
}

const logglyDateFormat = "2006-01-02T15:04:05.9999Z"

// getNowDate returns the current date as string in a valid format for loggly
func getNowDate() string {
	return time.Now().Format(logglyDateFormat)
}

// buildAndShipMessage creates the *logMessage to be send to loggly (adding current time) and ship it (send or add to the buffer)
func buildAndShipMessage(output string, messageType string, exit bool, data map[string]interface{}) {
	var formattedOutput string

	if data == nil {
		// Format message.
		formattedOutput = fmt.Sprintf("%v [%s] %s", getNowDate(), messageType, output)
	} else {
		// Format message.
		formattedOutput = fmt.Sprintf("%v [%s] %s %+v", getNowDate(), messageType, output, data)
	}

	if loggerSingleton.debugMode {
		fmt.Println(formattedOutput)
	}

	message := newMessage(getNowDate(), messageType, output, data)

	// Send message to loggly.
	ship(message)

	if exit {
		os.Exit(1)
	}
}

// newMessage creates a logMessage and return a pointer to it
func newMessage(timestamp string, level string, message string, data map[string]interface{}) *logMessage {
	formatedMessage := &logMessage{
		Timestamp: timestamp,
		Level:     level,
		Message:   message,
		Data:      data,
	}

	return formatedMessage
}

// ship depending on log configuration, it sends the message to loggly or add it to the buffer
func ship(message *logMessage) {
	// If bulk is set to true then ship on interval else ship the single log event.
	if loggerSingleton.bulk {
		go handleBulkLogMessage(message)
	} else {
		go handleLogMessage(message)
	}
}

// handleLogMessage immediately sends the given message to loggly
func handleLogMessage(message *logMessage) {
	requestBody, err := json.Marshal(message)

	if err != nil {
		fmt.Printf("There was an error marshalling log message: %s", err)
	}

	resp, err := http.Post(loggerSingleton.url, "text/plain", bytes.NewBuffer(requestBody))
	if err != nil {
		if loggerSingleton.debugMode {
			fmt.Printf("There was an error shipping the logs to loggy: %s", err)
		}
		return
	}

	if resp.StatusCode == 403 {
		if loggerSingleton.debugMode {
			fmt.Println("Token is invalid", resp.Status)
		}

	}

	if resp.StatusCode == 200 {
		if loggerSingleton.debugMode {
			fmt.Println("Log was shipped successfully", resp.Status)
		}
	}

	defer resp.Body.Close()

}

// handleBulkLogMessage adds the given message to the buffer, and send messages to loggly if max buffer size is achieved
func handleBulkLogMessage(message *logMessage) {
	var count int

	// Lock buffer from outside manipulation.
	loggerSingleton.Lock()

	loggerSingleton.buffer = append(loggerSingleton.buffer, message)

	count = len(loggerSingleton.buffer)

	// Unlock buffer from outside manipulation.
	loggerSingleton.Unlock()

	// Send buffer to loggly if the buffer size has been met.
	if count >= loggerSingleton.bufferSize {
		go flush()
	}

}

// flush sends the log messages in buffer to loggly. In case of errors, it puts back the messages to the buffer so can
// be sent the next time this is executed.
func flush() {
	loggerSingleton.Lock()
	messages := loggerSingleton.buffer
	loggerSingleton.buffer = make([]*logMessage, 0, loggerSingleton.bufferSize)
	loggerSingleton.Unlock()

	body := formatBulkMessages(messages)
	if len(body) == 0 {
		if loggerSingleton.debugMode {
			fmt.Println("No logs to send: Status OK")
		}
		return
	}

	resp, err := http.Post(loggerSingleton.url, "text/plain", bytes.NewBuffer([]byte(body)))
	if err != nil {
		if loggerSingleton.debugMode {
			fmt.Printf("There was an error shipping the logs to loggy: %s", err)
		}
		putMessagesBackToBuffer(messages)
		return
	}

	if resp.StatusCode == 403 {
		if loggerSingleton.debugMode {
			fmt.Println("Token is invalid", resp.Status)
		}
		putMessagesBackToBuffer(messages)
		return
	}

	if resp.StatusCode == 200 {
		if loggerSingleton.debugMode {
			fmt.Println("Logs were shipped successfully", resp.Status)
		}
	}

	defer resp.Body.Close()
}

// start sends periodically the buffer of log messages to loggly
func start() {
	for {
		time.Sleep(loggerSingleton.flushInterval)
		go flush()
	}
}

// tagList returns a string that contains all the tags to be send to loggly for these log messages
func tagList() string {
	return strings.Join(loggerSingleton.tags, ",")
}

// formatBulkMessages format all messages in given messagesBuffer to send them to loggly
func formatBulkMessages(messagesBuffer []*logMessage) string {
	var output string

	for _, m := range messagesBuffer {
		b, err := json.Marshal(m)

		if err != nil {
			fmt.Printf("There was an error marshalling buffer message: %s", err)
			continue
		}

		output += string(b) + "\n"
	}

	return output
}

// putMessagesBackToBuffer adds back messagesBuffer to loggerSingleton.buffer in case those were not sent successfully
func putMessagesBackToBuffer(messagesBuffer []*logMessage) {
	loggerSingleton.Lock()
	defer loggerSingleton.Unlock()
	loggerSingleton.buffer = append(loggerSingleton.buffer, messagesBuffer...)
}

var dataMessagesRegex = regexp.MustCompile(`@\w*%?[#+.\-\w]*\b`)

// formatDataMessages format messages replacing words that start with '@' with the string value of the element in 'a'.
// Also return data used in the message, to be used later in message to be sent to loggly.
// Example: format="example @Data message @Used", a=[1234, "text"] will return
// message="example 1234 message text" - data={"Data": 1234, "Used": "text"}
func formatDataMessages(format string, values ...interface{}) (message string, data map[string]interface{}) {
	i := 0
	data = make(map[string]interface{}, len(values))
	message = dataMessagesRegex.ReplaceAllStringFunc(format, func(expressionFound string) string {
		if i >= len(values) {
			return "[FORMAT ERROR]"
		}

		if len(expressionFound) <= 1 {
			return expressionFound
		}

		defaultValueFormat := "%v"
		var dataName string
		if strings.Contains(expressionFound, "%") {
			dataValueParts := strings.Split(expressionFound[1:], "%")
			dataName = dataValueParts[0]
			if len(dataValueParts) > 1 {
				defaultValueFormat = "%" + dataValueParts[1]
			}
		} else {
			dataName = expressionFound[1:]
		}

		dataValue := values[i]
		i++

		data[dataName] = dataValue

		return fmt.Sprintf("%v="+defaultValueFormat, dataName, dataValue)
	})
	return message, data
}
