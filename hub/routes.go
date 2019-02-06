package hub

func (hub *Hub) routes() {
	hub.router.HandleFunc("/api/rotators", hub.rotatorsHandler).Methods("GET")
	hub.router.HandleFunc("/api/rotator/{rotator}", hub.rotatorHandler).Methods("GET")
	hub.router.HandleFunc("/api/rotator/{rotator}/azimuth", hub.azimuthHandler)
	hub.router.HandleFunc("/api/rotator/{rotator}/elevation", hub.elevationHandler)
	hub.router.HandleFunc("/api/rotator/{rotator}/stop", hub.stopHandler)
	hub.router.HandleFunc("/api/rotator/{rotator}/stop_azimuth", hub.stopAzimuthHandler)
	hub.router.HandleFunc("/api/rotator/{rotator}/stop_elevation", hub.stopElevationHandler)
	hub.router.HandleFunc("/ws", hub.wsHandler)
	hub.router.PathPrefix("/").Handler(hub.fileServer)
}
