package mylog

import (
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	logrus "github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	//"log"
)

var (
    // Log
    Logger = logrus.New()
    //LogInfo = log.New(os.Stdout, "\u001b[37m", log.LstdFlags)
    //LogWarning = log.New(os.Stdout, "\u001b[33mWARNING: ", log.LstdFlags)
    //LogError = log.New(os.Stdout, "\u001b[35m Error: \u001B[31m", log.LstdFlags)
    //LogDebug = log.New(os.Stdout, "\u001b[36mDebug: ", log.LstdFlags)
)

func InitLogger() {
    println("Initializing Logger")
    formatter := &prefixed.TextFormatter{
        TimestampFormat : "2006-01-02  15:04:05",
        ForceColors :true,
        FullTimestamp:true,
        ForceFormatting: true,
    }
    // if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
    //     println("No TTY detected - Disabling colors")
    //     formatter.DisableColors = true;
    //     formatter.ForceColors = false;
    // }
    loglevel := Getenv("LOG_LEVEL", "INFO")
    loglevelInt, err := logrus.ParseLevel(loglevel)
    if err != nil {logrus.Panic(err)}
    
    Logger = &logrus.Logger{
        Out:   colorable.NewColorableStderr(),
        Level: loglevelInt,
        Formatter: formatter,
    }
}

// Attempts to resolve an environment variable, 
//  with a default value if it's empty
func Getenv(key, fallback string) string {
    value := os.Getenv(key)
    if len(value) == 0 {
        return fallback
    }
    return value
}
