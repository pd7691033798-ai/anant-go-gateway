rawError := r.URL.Query().Get("error_log")

		ticket := hub.AppSupport.IngestAndAutoResolve("IN_APP_CRASH_HOOK", parentID, rawError)
		hub.Brain.ProcessFeedbackAndFinance(string(ticket.Type), rawError)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ticket)
	})

	// एडमिन स्टेटस
	http.HandleFunc("/api/v1/admin/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"system":        "Anant Abhyas Ultra Core",
			"cluster_state": "ACTIVE",
			"auto_pay":      "ENFORCED_MANDATE_ONLY",
			"active_tracks": []string{"NAVODAYA", "SAINIK_SCHOOL", "NDA", "IIT_JEE"},
		})
	})

	// 5. सर्वर स्टार्टअप
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 अनंत अभ्यास क्लस्टर पोर्ट :%s पर पूर्णतः सक्रिय है...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
