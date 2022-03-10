package mylog

import (
    logrus "github.com/sirupsen/logrus"
    // prefixed "github.com/x-cray/logrus-prefixed-formatter"
    "github.com/mattn/go-colorable"
    "os"
    "bytes"
    "fmt"
    "runtime/debug"
)

var (
    // Log
    Logger *logrus.Logger;

    // Old logger: import("log")
    //LogInfo = log.New(os.Stdout, "\u001b[37m", log.LstdFlags)
    //LogWarning = log.New(os.Stdout, "\u001b[33mWARNING: ", log.LstdFlags)
    //LogError = log.New(os.Stdout, "\u001b[35m Error: \u001B[31m", log.LstdFlags)
    //LogDebug = log.New(os.Stdout, "\u001b[36mDebug: ", log.LstdFlags)

    enableColorsStdout = true

    levelList = [] string{
        "PANIC",
        "FATAL",
        "ERROR",
        "WARN",
        "INFO",
        "DEBUG",
        "TRACE",
    }

)

type MyFormatter struct {}

func InitLogger() {
    println("Initializing Logger")
    // colorable.EnableColorsStdout(&enableColorsStdout)
    logrus.SetReportCaller(true)
    Logger = logrus.New()
    // formatter := &prefixed.TextFormatter{
    //     TimestampFormat : "2006-01-02 15:04:05",
    //     ForceColors :true,
    //     FullTimestamp:true,
    //     ForceFormatting: true,
    // }
    // if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
    //     println("No TTY detected - Disabling stdout colors")
    //     formatter.DisableColors = true;
    //     formatter.ForceColors = false;
    // }
    // formatter.SetColorScheme(&prefixed.ColorScheme{
    //     PrefixStyle: "blue+b",
    // })
    Logger = &logrus.Logger{
        Out:   os.Stderr,
        Level: logrus.InfoLevel,
        Formatter: &MyFormatter{},
    }
    Logger.SetOutput(colorable.NewColorableStdout())
    
}


func (mf *MyFormatter) Format(entry *logrus.Entry) ([]byte, error){
    var b *bytes.Buffer
    if entry.Buffer != nil {
        b = entry.Buffer
    } else {
        b = &bytes.Buffer{}
    }
    level := levelList[int(entry.Level)]
    if entry.Level < 2 {  // Fatal & Panic
        stack := debug.Stack()
        b.WriteString(fmt.Sprintf("%s\n",
        stack))
    }
    b.WriteString(fmt.Sprintf("%s  [%s] : %s\n",
        entry.Time.Format("2006-01-02 15:04:05"), 
        level, 
        entry.Message))

    return b.Bytes(), nil
}