package healthplanet

type HealthPlanetAuth interface {
    GetToken() (string, error)
}
