package main

import (
    "log"
    "net/http"
    "os"
    "time"
)

func StartKeepAlive() {
    appURL := os.Getenv("APP_URL")
    if appURL == "" {
        log.Println("Keep-Alive: APP_URL सेट नहीं है, सेल्फ-पिंग बंद है।")
        return
    }

    ticker := time.NewTicker(10 * time.Minute)
    go func() {
        for range ticker.C {
            resp, err := http.Get(appURL)
            if err != nil {
                log.Printf("Keep-Alive Ping विफल: %v\n", err)
                continue
            }
            resp.Body.Close()
            log.Printf("Keep-Alive Ping सफल: सर्वर जाग रहा है (Status: %s)\n", resp.Status)
        }
    }()
}
