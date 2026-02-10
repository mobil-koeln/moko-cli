package api

const (
	// BaseURL is the base URL for the bahn.de API
	BaseURL = "https://www.bahn.de/web/api"

	// EndpointDepartures returns departures at a station
	// Required params: datum, zeit, ortExtId, ortId, mitVias, maxVias, verkehrsmittel[]
	EndpointDepartures = "/reiseloesung/abfahrten"

	// EndpointArrivals returns arrivals at a station
	// Required params: datum, zeit, ortExtId, ortId, mitVias, maxVias, verkehrsmittel[]
	EndpointArrivals = "/reiseloesung/ankuenfte"

	// EndpointLocations searches for stations by name
	// Required params: suchbegriff, typ, limit
	EndpointLocations = "/reiseloesung/orte"

	// EndpointNearby searches for stations by coordinates
	// Required params: lat, long, radius, maxNo
	EndpointNearby = "/reiseloesung/orte/nearby"

	// EndpointJourney returns journey/trip details
	// Required params: journeyId, poly
	EndpointJourney = "/reiseloesung/fahrt"

	// EndpointFormation returns train carriage formation
	// Required params: administrationId, category, date, evaNumber, number, time
	EndpointFormation = "/reisebegleitung/wagenreihung/vehicle-sequence"
)

// ModesOfTransit contains all supported transport modes
var ModesOfTransit = []string{
	"ICE",
	"EC_IC",
	"IR",
	"REGIONAL",
	"SBAHN",
	"BUS",
	"SCHIFF",
	"UBAHN",
	"TRAM",
	"ANRUFPFLICHTIG",
}
