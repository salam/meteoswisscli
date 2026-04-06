package source

import "testing"

func TestMeteoSwissAttribution(t *testing.T) {
	want := "Quelle: MeteoSchweiz; Source: MétéoSuisse; Fonte: MeteoSvizzera; Source: MeteoSwiss"
	if MeteoSwiss != want {
		t.Errorf("MeteoSwiss = %q, want %q", MeteoSwiss, want)
	}
}

func TestSLFAttribution(t *testing.T) {
	want := "Quelle: SLF/WSL; Source: SLF/WSL"
	if SLF != want {
		t.Errorf("SLF = %q, want %q", SLF, want)
	}
}
