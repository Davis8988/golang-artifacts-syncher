package mylog

import (
    logrus "github.com/sirupsen/logrus"
    prefixed "github.com/x-cray/logrus-prefixed-formatter"
    "github.com/mattn/go-colorable"
    //"log"
)

var (
    // Log
    Logger = logrus.New()
    //LogInfo = log.New(os.Stdout, "\u001b[37m", log.LstdFlags)
    //LogWarning = log.New(os.Stdout, "\u001b[33mWARNING: ", log.LstdFlags)
    //LogError = log.New(os.Stdout, "\u001b[35m Error: \u001B[31m", log.LstdFlags)
    //LogDebug = log.New(os.Stdout, "\u001b[36mDebug: ", log.LstdFlags)
    levelList = map[string] int{
        "PANIC",
        "FATAL",
        "ERROR",
        "WARN",
        "INFO",
        "DEBUG",
        "TRACE",
    }

    := map[int]string{
     
        90: "Dog",
        91: "Cat",
        92: "Cow",
        93: "Bird",
        94: "Rabbit",
}
)

func InitLogger() {
    println("Initializing Logger")
    loglevel := Getenv("LOG_LEVEL", "INFO")
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

    Logger = &logrus.Logger{
        Out:   colorable.NewColorableStderr(),
        Level: logrus.InfoLevel,
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
