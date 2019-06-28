package hub

func (hub *Hub) routes() {
	// API v1.0
	hub.router.HandleFunc("/api/v1.0/rotators", hub.rotatorsHandler).Methods("GET")
	hub.router.HandleFunc("/api/v1.0/rotator/{rotator}", hub.rotatorHandler).Methods("GET")
	hub.router.HandleFunc("/api/v1.0/rotator/{rotator}/azimuth", hub.azimuthHandler)
	hub.router.HandleFunc("/api/v1.0/rotator/{rotator}/elevation", hub.elevationHandler)
	hub.router.HandleFunc("/api/v1.0/rotator/{rotator}/stop", hub.stopHandler)
	hub.router.HandleFunc("/api/v1.0/rotator/{rotator}/stop_azimuth", hub.stopAzimuthHandler)
	hub.router.HandleFunc("/api/v1.0/rotator/{rotator}/stop_elevation", hub.stopElevationHandler)

	hub.router.HandleFunc("/ws", hub.wsHandler)
	hub.router.PathPrefix("/").Handler(hub.fileServer)
}
