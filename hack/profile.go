package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/go-openapi/spec"
	"github.com/go-openapi/swag"
	"github.com/pkg/profile"
)

func main() {
	server := httptest.NewServer(http.FileServer(http.Dir("fixtures/azure")))
	defer server.Close()

	//defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
	//defer profile.Start(profile.MemProfile, profile.ProfilePath(".")).Stop()
	defer profile.Start(profile.MemProfile, profile.MemProfileAllocs(), profile.ProfilePath(".")).Stop()
	for i := 0; i < 10; i++ {
		basePath := server.URL + "/publicIpAddress.json"
		doc, err := swag.LoadFromFileOrHTTP(basePath)
		if err != nil {
			log.Fatalf("load: %v", err)
		}

		opts := &spec.ExpandOptions{
			RelativeBase: basePath,
		}

		sp := new(spec.Swagger)

		err = json.Unmarshal(doc, sp)
		if err != nil {
			log.Fatalf("unmarshal: %v", err)
		}

		err = spec.ExpandSpec(sp, opts)
		if err != nil {
			log.Fatalf("expand: %v", err)
		}
	}
	/*
		b, err := json.MarshalIndent(sp, "", " ")
		if err != nil {
			log.Fatalf("marshal: %v", err)
		}
		fmt.Println(string(b))
	*/
}
