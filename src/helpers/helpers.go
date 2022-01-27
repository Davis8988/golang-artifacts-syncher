package helpers

import (
	"os"
	"sync"
)

// Attempts to resolve an environment variable, 
//  with a default value if it's empty
func Getenv(key, fallback string) string {
    value := os.Getenv(key)
    if len(value) == 0 {
        return fallback
    }
    return value
}

func IsStrArrayEmpty(arrToCheck []string) bool {
	return len(arrToCheck) == 0
}

func loadStringArrValueFromSynchedMap(someMap sync.Map, key string) [] string {
    currentInterfaceValue, _ := someMap.Load(key)
    var currentStrArr []string = currentInterfaceValue.([]string)
    return currentStrArr
}

