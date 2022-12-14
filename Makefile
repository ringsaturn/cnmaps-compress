build:
	go build

gen:
	git submodule update --init --recursive
	make build
	cd cnmaps/cnmaps/data/geojson.min/administrative/amap/;mkdir land-reduce-json;mkdir land-reduce-pb
	./cnmaps-compress
	cd cnmaps/cnmaps/data/geojson.min/administrative/amap/land;mv *.json ../land-reduce-json
	cd cnmaps/cnmaps/data/geojson.min/administrative/amap/land;mv *.pb ../land-reduce-pb
