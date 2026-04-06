package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseVQHA80(t *testing.T) {
	csv := "Station/Location;Date;tre200s0;rre150z0;sre000z0;gre000z0;ure200s0;tde200s0;dkl010z0;fu3010z0;fu3010z1;prestas0;pp0qffs0;pp0qnhs0;ppz850s0;ppz700s0;dv1towz0;fu3towz0;fu3towz1;ta1tows0;uretows0;tdetows0\nTAE;202604051910;15.2;0.0;10;245;55;6.3;250;12.6;22.3;965.1;1015.2;1015.0;-;-;-;-;-;-;-;-\nSMA;202604051910;18.1;0.0;10;180;42;5.1;210;8.4;18.7;955.8;1013.8;1013.6;-;-;230;14.0;28.1;17.2;40;3.6\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(csv))
	}))
	defer server.Close()

	c := &Client{http: &http.Client{}, baseURL: server.URL, lang: "de"}
	stations, err := c.GetCurrentMeasurements(server.URL + "/test.csv")
	if err != nil {
		t.Fatalf("GetCurrentMeasurements error = %v", err)
	}
	if len(stations) != 2 {
		t.Fatalf("stations = %d, want 2", len(stations))
	}
	if stations[0].Station != "TAE" {
		t.Errorf("station[0] = %q, want TAE", stations[0].Station)
	}
	if stations[0].Temperature != "15.2" {
		t.Errorf("temperature = %q, want 15.2", stations[0].Temperature)
	}
}
