package log

import (
	"testing"
	"time"
)

func TestSetupSingleLogger(t *testing.T) {
	SetupLogger("yourlogglytoken", 0, []string{"test"}, false, true)

	Infoln("This is an info statement.")
}

func TestSetupBulkLogger(t *testing.T) {
	SetupLogger("yourlogglytoken", 0, []string{"test"}, true, true)

	Infoln("This is an info statement 1.")
	Infoln("This is an info statement 2.")
	Infoln("This is an info statement 3.")
	Infoln("This is an info statement 4.")

	time.Sleep(11 * time.Second)
}

func TestInfod(t *testing.T) {
	SetupLogger("yourlogglytoken", 0, []string{"test"}, false, true)

	data := map[string]interface{}{
		"Name": "Logan",
		"Kind": "Log",
	}

	Infod("This is some data", data)

	time.Sleep(3 * time.Second)
}

func TestDebugln(t *testing.T) {
	Debugln("This is a debug statement.")
}

func TestDebugf(t *testing.T) {
	Debugf("This is a debug statement %d.", 10000)
}

func TestInfoln(t *testing.T) {
	Infoln("This is an info statement.")
}

func TestInfof(t *testing.T) {
	Infof("This is an info statement %d.", 10000)
}

func TestWarnln(t *testing.T) {
	Warnln("This is a warning.")
}

func TestWarnf(t *testing.T) {
	Warnf("This is a warning %d.", 10000)
}

func TestErrorln(t *testing.T) {
	Errorln("This is an error.")
}

func TestErrorf(t *testing.T) {
	Errorf("This is an error %d.", 10000)
}

func TestFatalln(t *testing.T) {
	Fatalln("This is an error.")
}

func TestFatalf(t *testing.T) {
	Fatalf("This is an error %d.", 10000)
}

func TestFormatDataMessagesSimpleMessage(t *testing.T) {
	message, data := formatDataMessages("@UserId unlocked @GachaId", "<USERID>", 5)
	expectedMessage := "UserId=<USERID> unlocked GachaId=5"
	if message != expectedMessage {
		t.Errorf("Expected %s, got %s", expectedMessage, message)
	}

	if len(data) != 2 {
		t.Errorf("Expected data len = 2, got %v", len(data))
	}

	if data["UserId"] != "<USERID>" {
		t.Errorf("Expected data['UserId'] = 1, got %v", data["UserId"])
	}

	if data["GachaId"] != 5 {
		t.Errorf("Expected data['GachaId'] = 5, got %v", data["GachaId"])
	}
}

func TestFormatDataMessageWithNoReplaces(t *testing.T) {
	noReplacesMessage := "@ nothing"
	message, data := formatDataMessages(noReplacesMessage, "NO")
	if message != "@ nothing" {
		t.Errorf("Expected %s, got %s", noReplacesMessage, message)
	}

	if len(data) != 0 {
		t.Errorf("Expected data len = 0, got %v - data=%v", len(data), data)
	}
}

func TestFormatDataMessagesComplexMessagesAndFormats(t *testing.T) {
	type testStruct struct {
		Name string
		Age  int
	}

	complexData := testStruct{
		Name: "name",
		Age:  2,
	}

	message, data := formatDataMessages("testStruct @Struct", complexData)
	expectedMessage := "testStruct Struct={name 2}"
	if message != expectedMessage {
		t.Errorf("Expected %s, got %s", expectedMessage, message)
	}
	if len(data) != 1 {
		t.Errorf("Expected data len = 1, got %v", len(data))
	}

	if data["Struct"] != complexData {
		t.Errorf("Expected data['Struct'] = %+v, got %+v", complexData, data["Struct"])
	}

	message, data = formatDataMessages("testStruct @Struct%+v", complexData)
	expectedMessage = "testStruct Struct={Name:name Age:2}"
	if message != expectedMessage {
		t.Errorf("Expected %s, got %s", expectedMessage, message)
	}

	message, data = formatDataMessages("testFloat @Float%.2f", 1.23456)
	expectedMessage = "testFloat Float=1.23"
	if message != expectedMessage {
		t.Errorf("Expected %s, got %s", expectedMessage, message)
	}

	message, data = formatDataMessages("@string%-6s", "cafe")
	expectedMessage = "string=cafe  "
	if message != expectedMessage {
		t.Errorf("Expected %s, got %s", expectedMessage, message)
	}

	message, data = formatDataMessages("@int%04d", 15)
	expectedMessage = "int=0015"
	if message != expectedMessage {
		t.Errorf("Expected %s, got %s", expectedMessage, message)
	}

}
