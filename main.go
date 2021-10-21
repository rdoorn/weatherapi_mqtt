package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/rdoorn/gohelper/mqtthelper"
)

const mqttClientID = "weatherapi_mqtt"

type TelemetryMQTTStatus struct {
	Time    *int64  `json:"time"`
	TimeStr *string `json:"time_string"`
	Summary *string `json:"summary"`
	// Icon:clear-day
	SunriseTime  int64   `json:"sunrise_time"`
	SunsetTime   int64   `json:"sunset_time"`
	SunriseTimeH float64 `json:"sunrise_time_h"`
	SunsetTimeH  float64 `json:"sunset_time_h"`
	//SunsetTime:0
	PrecipIntensity *float64 `json:"rain_intensity"`
	//PrecipIntensityMax:0
	//PrecipIntensityMaxTime:0
	//PrecipProbability *float64 `json:"rain_posibility"`
	//PrecipType:  rain|
	//PrecipAccumulation:0
	Temperature *float64 `json:"temperature"`
	//TemperatureMin:0
	//TemperatureMinTime:0
	//TemperatureMax:0
	//TemperatureMaxTime:0
	ApparentTemperature *float64 `json:"apparent_temperature"`
	//ApparentTemperatureMin:0
	//ApparentTemperatureMinTime:0
	//ApparentTemperatureMax:0
	//ApparentTemperatureMaxTime:0
	//NearestStormBearing  *float64 `json:"nearest_storm_bearing"`
	//NearestStormDistance *float64 `json:"nearest_storm_distance"`

	//DewPoint         *float64 `json:"dew_point"`
	WindSpeed        *float64 `json:"wind_speed"`
	WindGust         *float64 `json:"wind_gust"`
	WindBearing      *int64   `json:"wind_bearing"`
	CloudCover       *int64   `json:"cloud_cover"`
	Humidity         *int64   `json:"humidity"`
	Pressure         *float64 `json:"pressure"`
	Visibility       *float64 `json:"visibility"`
	Ozone            *float64 `json:"ozone"`
	CarbonOxide      *float64 `json:"carbon_oxide"`
	NitrogenOxide    *float64 `json:"nitrogen_oxide"`
	SulphurDioxide   *float64 `json:"sulphur_dioxide"`
	PM2_5            *float64 `json:"pm2_5"`
	PM10             *float64 `json:"pm10"`
	MoonPhase        *string  `json:"moon_phase"`
	MoonIllumination *int     `json:"moon_illumination"`
	UVIndex          *float64 `json:"uv_index"`
	//UVIndexTime float64
}

type Handler struct {
	mqtt           *mqtthelper.Handler
	weatherapiAPI  string
	weatherapiLong string
	weatherapiLat  string
}

func (n TelemetryMQTTStatus) String() string {
	b, err := json.Marshal(n)
	if err != nil {
		return err.Error()
	}

	return string(b)
}

func main() {
	weatherapiAPI, ok := os.LookupEnv("WEATHERAPI_API")
	if !ok {
		panic("missing environment key: WEATHERAPI_API")
	}
	weatherapiLong, ok := os.LookupEnv("WEATHERAPI_LONG")
	if !ok {
		panic("missing environment key: WEATHERAPI_LONG")
	}
	weatherapiLat, ok := os.LookupEnv("WEATHERAPI_LAT")
	if !ok {
		panic("missing environment key: WEATHERAPI_LAT")
	}

	h := Handler{
		weatherapiAPI:  weatherapiAPI,
		weatherapiLong: weatherapiLong,
		weatherapiLat:  weatherapiLat,
		mqtt:           mqtthelper.New(),
	}

	// run solaredge get on a timer
	ticker := time.NewTicker(5 * time.Minute)
	log.Printf("starting poll")
	h.poll()

	// loop till exit
	sigterm := make(chan os.Signal, 10)
	signal.Notify(sigterm, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-sigterm:
			log.Printf("Program killed by signal!")
			ticker.Stop()
			return
		case <-ticker.C:
			if err := h.poll(); err != nil {
				log.Printf("poll status failed: %s", err)
			}
		}
	}

}

func (h *Handler) poll() error {
	f := &CurrentJsonResponse{}
	err := getJson(fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&q=%s,%s&aqi=yes", h.weatherapiAPI, h.weatherapiLat, h.weatherapiLong), f)
	if err != nil {
		log.Fatal(err)
	}

	year, month, day := time.Now().Date()

	a := &AstronomyJsonResponse{}
	err = getJson(fmt.Sprintf("http://api.weatherapi.com/v1/astronomy.json?key=%s&q=%s,%s&dt=%d-%02d-%02d", h.weatherapiAPI, h.weatherapiLat, h.weatherapiLong, year, month, day), a)
	if err != nil {
		log.Fatal(err)
	}

	/*
		f, err := forecast.Get(h.weatherapiAPI, h.weatherapiLat, h.weatherapiLong, "now", forecast.CA, forecast.English)
		if err != nil {
			log.Fatal(err)
		}
	*/
	currentString := "current"

	moonIllumination, _ := strconv.Atoi(*a.Astronomy.Astro.MoonIllumination)

	log.Printf("forecast: %+v", f)
	current := TelemetryMQTTStatus{
		Time:            f.Current.LastUpdatedEpoch,
		TimeStr:         &currentString,
		Summary:         f.Current.Condition.Text,
		PrecipIntensity: f.Current.PressureIn,
		//PrecipProbability:    f.Current.PrecipProbability,
		Temperature:         f.Current.TempC,
		ApparentTemperature: f.Current.FeelslikeC,
		//NearestStormBearing:  f.Current.NearestStormBearing,
		//NearestStormDistance: f.Current.NearestStormDistance,
		//DewPoint:             f.Current.DewPoint,
		WindSpeed:        f.Current.WindKph,
		WindGust:         f.Current.GustKph,
		WindBearing:      f.Current.WindDegree,
		CloudCover:       f.Current.Cloud,
		Humidity:         f.Current.Humidity,
		Pressure:         f.Current.PressureMb,
		Visibility:       f.Current.VisKm,
		CarbonOxide:      f.Current.AirQuality.Co,
		Ozone:            f.Current.AirQuality.O3,
		NitrogenOxide:    f.Current.AirQuality.No2,
		SulphurDioxide:   f.Current.AirQuality.So2,
		PM2_5:            f.Current.AirQuality.Pm2_5,
		PM10:             f.Current.AirQuality.Pm10,
		MoonPhase:        a.Astronomy.Astro.MoonPhase,
		MoonIllumination: &moonIllumination,
		UVIndex:          f.Current.Uv,

		SunsetTime:   TimeToEpoch(a.Astronomy.Astro.Sunset).Unix(),
		SunriseTime:  TimeToEpoch(a.Astronomy.Astro.Sunrise).Unix(),
		SunsetTimeH:  float64(float64(TimeToEpoch(a.Astronomy.Astro.Sunset).Hour()) + float64(float64(TimeToEpoch(a.Astronomy.Astro.Sunset).Minute())/float64(60))),
		SunriseTimeH: float64(float64(TimeToEpoch(a.Astronomy.Astro.Sunrise).Hour()) + float64(float64(TimeToEpoch(a.Astronomy.Astro.Sunrise).Minute())/float64(60))),
	}
	h.mqtt.Publish("weatherapi/out", 0, false, current.String())

	/*
		for i := 0; i < 24 && i < len(f.Hourly.Data); i++ {
			next := TelemetryMQTTStatus{
				Time:                 f.Hourly.Data[i].Time,
				TimeStr:              fmt.Sprintf("%dh", i+1),
				Summary:              f.Hourly.Data[i].Summary,
				PrecipIntensity:      f.Hourly.Data[i].PrecipIntensity,
				PrecipProbability:    f.Hourly.Data[i].PrecipProbability,
				Temperature:          f.Hourly.Data[i].Temperature,
				ApparentTemperature:  f.Hourly.Data[i].ApparentTemperature,
				NearestStormBearing:  f.Hourly.Data[i].NearestStormBearing,
				NearestStormDistance: f.Hourly.Data[i].NearestStormDistance,
				DewPoint:             f.Hourly.Data[i].DewPoint,
				WindSpeed:            f.Hourly.Data[i].WindSpeed,
				WindGust:             f.Hourly.Data[i].WindGust,
				WindBearing:          f.Hourly.Data[i].WindBearing,
				CloudCover:           f.Hourly.Data[i].CloudCover,
				Humidity:             f.Hourly.Data[i].Humidity,
				Pressure:             f.Hourly.Data[i].Pressure,
				Visibility:           f.Hourly.Data[i].Visibility,
				Ozone:                f.Hourly.Data[i].Ozone,
				MoonPhase:            f.Hourly.Data[i].MoonPhase,
				UVIndex:              f.Hourly.Data[i].UVIndex,
			}
			log.Printf("sending temperature for %dh: %f", i+1, f.Hourly.Data[i].Temperature)
			h.mqtt.Publish(mqttClientID, "weatherapi/out", 0, false, next.String())
		}

		if len(f.Daily.Data) > 0 {
			current := TelemetryMQTTStatus{
				Time:                 f.Daily.Data[0].Time,
				TimeStr:              "daily",
				Summary:              f.Daily.Data[0].Summary,
				PrecipIntensity:      f.Daily.Data[0].PrecipIntensity,
				PrecipProbability:    f.Daily.Data[0].PrecipProbability,
				Temperature:          f.Daily.Data[0].Temperature,
				ApparentTemperature:  f.Daily.Data[0].ApparentTemperature,
				NearestStormBearing:  f.Daily.Data[0].NearestStormBearing,
				NearestStormDistance: f.Daily.Data[0].NearestStormDistance,
				DewPoint:             f.Daily.Data[0].DewPoint,
				WindSpeed:            f.Daily.Data[0].WindSpeed,
				WindGust:             f.Daily.Data[0].WindGust,
				WindBearing:          f.Daily.Data[0].WindBearing,
				CloudCover:           f.Daily.Data[0].CloudCover,
				Humidity:             f.Daily.Data[0].Humidity,
				Pressure:             f.Daily.Data[0].Pressure,
				Visibility:           f.Daily.Data[0].Visibility,
				Ozone:                f.Daily.Data[0].Ozone,
				MoonPhase:            f.Daily.Data[0].MoonPhase,
				SunsetTime:           f.Daily.Data[0].SunsetTime,
				SunriseTime:          f.Daily.Data[0].SunriseTime,
				SunsetTimeH:          float64(float64(time.Unix(f.Daily.Data[0].SunsetTime, 0).Hour()) + float64(float64(time.Unix(f.Daily.Data[0].SunsetTime, 0).Minute())/float64(60))),
				SunriseTimeH:         float64(float64(time.Unix(f.Daily.Data[0].SunriseTime, 0).Hour()) + float64(float64(time.Unix(f.Daily.Data[0].SunriseTime, 0).Minute())/float64(60))),
				UVIndex:              f.Daily.Data[0].UVIndex,
			}
			log.Printf("sending daily temperature: %f", f.Daily.Data[0].Temperature)
			h.mqtt.Publish(mqttClientID, "weatherapi/out", 0, false, current.String())
		}
	*/

	return nil
}

// 	s := "08:17 AM"
func TimeToEpoch(s *string) time.Time {
	year, month, day := time.Now().Date()
	s2 := fmt.Sprintf("%d-%02d-%02d %s", year, month, day, *s)

	layout := "2006-01-02 15:04 PM"
	t, err := time.Parse(layout, s2)
	if err != nil {
		log.Printf("err parsing %s: %s", s2, err)
	}
	return t

}
