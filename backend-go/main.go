package main

import (
 "log"
 "net/http"
 "os"

 "terrafertil/backend-go/internal/api"
 "terrafertil/backend-go/internal/store"
)

func main() {
 port := getenv("PORT", "8080")
 aiServiceURL := getenv("AI_SERVICE_URL", "http://localhost:8001")
 databasePath := getenv("DATABASE_PATH", "./terrafertil.db")

 db, err := store.OpenSQLite(databasePath)
 if err != nil {
  log.Fatalf("could not open database: %v", err)
 }
 defer db.Close()

 waitlistStore, err := store.NewWaitlistStore(db)
 if err != nil {
  log.Fatalf("could not initialize waitlist store: %v", err)
 }

 mux := http.NewServeMux()
 server := api.NewServer(aiServiceURL, waitlistStore)
 server.RegisterRoutes(mux)

 handler := withCORS(mux)

 log.Printf("backend-go listening on :%s", port)
 log.Printf("using ai service: %s", aiServiceURL)
 log.Printf("using sqlite database: %s", databasePath)
 if err := http.ListenAndServe(":"+port, handler); err != nil {
  log.Fatalf("server stopped: %v", err)
 }
}

func withCORS(next http.Handler) http.Handler {
 return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Access-Control-Allow-Origin", "*")
  w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
  w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

  if r.Method == http.MethodOptions {
   w.WriteHeader(http.StatusNoContent)
   return
  }
  next.ServeHTTP(w, r)
 })
}

func getenv(key string, fallback string) string {
 if v := os.Getenv(key); v != "" {
  return v
 }
 return fallback
}
