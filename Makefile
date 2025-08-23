iti_testp1 iti_testp2 opendnp3_test1 opendnp3_test2 opendnp3_test3 opendnp3_test4: 
	@echo "Running on $(@)"
	@go run . examples/$(@).pcap | grep "error" -B 1 || echo "No errors"

test: iti_testp1 iti_testp2 opendnp3_test1 opendnp3_test2 opendnp3_test3 opendnp3_test4