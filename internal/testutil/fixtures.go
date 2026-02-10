package testutil

// Sample JSON responses for API testing

// SampleDepartureResponse is a minimal valid departure board response
const SampleDepartureResponse = `{
	"entries": [
		{
			"journeyId": "1|123456|0|80|1012024",
			"bahnhofsId": "8000105",
			"terminus": "München Hbf",
			"gleis": "7",
			"ezGleis": "8",
			"zeit": "2024-01-01T14:30:00",
			"ezZeit": "2024-01-01T14:32:00",
			"ueber": ["Mannheim", "Stuttgart"],
			"verkehrmittel": {
				"kurzText": "ICE",
				"mittelText": "ICE 123",
				"langText": "ICE 123 nach München",
				"name": "ICE 123"
			},
			"meldungen": []
		}
	]
}`

// SampleArrivalResponse is a minimal valid arrival board response
const SampleArrivalResponse = `{
	"entries": [
		{
			"journeyId": "1|654321|0|80|1012024",
			"bahnhofsId": "8000105",
			"terminus": "Frankfurt(Main)Hbf",
			"gleis": "12",
			"zeit": "2024-01-01T14:30:00",
			"verkehrmittel": {
				"kurzText": "ICE",
				"mittelText": "ICE 456",
				"langText": "ICE 456 nach Frankfurt",
				"name": "ICE 456"
			},
			"meldungen": []
		}
	]
}`

// SampleLocationResponse is a minimal valid location search response
const SampleLocationResponse = `[
	{
		"name": "Frankfurt(Main)Hbf",
		"evaId": "8000105",
		"id": "A=1@O=Frankfurt(Main)Hbf@X=8663785@Y=50107145@U=80@L=8000105@",
		"type": "STATION",
		"latLng": {
			"latitude": 50.107145,
			"longitude": 8.663785
		}
	},
	{
		"name": "Frankfurt(Main) Süd",
		"evaId": "8002041",
		"id": "A=1@O=Frankfurt(Main) Süd@X=8663395@Y=50099342@U=80@L=8002041@",
		"type": "STATION",
		"latLng": {
			"latitude": 50.099342,
			"longitude": 8.663395
		}
	}
]`

// SampleJourneyResponse is a minimal valid journey detail response
const SampleJourneyResponse = `{
	"journey": {
		"id": "1|123456|0|80|1012024",
		"date": "2024-01-01",
		"stops": [
			{
				"station": {
					"name": "Frankfurt(Main)Hbf",
					"evaId": "8000105",
					"id": "A=1@O=Frankfurt(Main)Hbf@X=8663785@Y=50107145@U=80@L=8000105@"
				},
				"departure": {
					"scheduledTime": "14:30",
					"time": "14:32"
				},
				"platform": "7"
			},
			{
				"station": {
					"name": "Mannheim Hbf",
					"evaId": "8000244",
					"id": "A=1@O=Mannheim Hbf@X=8469343@Y=49479557@U=80@L=8000244@"
				},
				"arrival": {
					"scheduledTime": "15:15",
					"time": "15:15"
				},
				"departure": {
					"scheduledTime": "15:17",
					"time": "15:17"
				},
				"platform": "5"
			},
			{
				"station": {
					"name": "München Hbf",
					"evaId": "8000261",
					"id": "A=1@O=München Hbf@X=11558339@Y=48140229@U=80@L=8000261@"
				},
				"arrival": {
					"scheduledTime": "17:45",
					"time": "17:50"
				},
				"platform": "18"
			}
		]
	}
}`

// SampleFormationResponse is a minimal valid train formation response
const SampleFormationResponse = `{
	"data": {
		"istformation": {
			"allFahrzeuggruppe": [
				{
					"fahrzeugnummer": "123456",
					"verkehrsmitteltyp": "ICE",
					"allFahrzeug": [
						{
							"fahrzeugnummer": "91 80 7 Apekzf 411.8 D-DB",
							"kategorie": "LOK",
							"wagenordnungsnummer": "1",
							"positioningruppe": "1"
						},
						{
							"fahrzeugnummer": "91 80 7 Bpmbkz 861.2 D-DB",
							"kategorie": "ERSTEKLASSE",
							"wagenordnungsnummer": "2",
							"positioningruppe": "2",
							"ausstattung": ["RUHE", "WIFI"]
						}
					]
				}
			]
		}
	}
}`

// SampleEmptyResponse is an empty JSON response
const SampleEmptyResponse = `{}`

// SampleErrorResponse is a sample error response
const SampleErrorResponse = `{
	"error": {
		"code": "STATION_NOT_FOUND",
		"message": "Station not found"
	}
}`
