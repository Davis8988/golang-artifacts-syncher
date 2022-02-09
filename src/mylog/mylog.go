package mylog

import (
    "log"
    "os"
)

var (
    // Log
    LogInfo = log.New(os.Stdout, "\u001b[37m", log.LstdFlags)
    LogWarning = log.New(os.Stdout, "\u001b[33mWARNING: ", log.LstdFlags)
    LogError = log.New(os.Stdout, "\u001b[35m Error: \u001B[31m", log.LstdFlags)
    LogDebug = log.New(os.Stdout, "\u001b[36mDebug: ", log.LstdFlags)
)
