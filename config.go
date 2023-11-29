package main

var Config = struct {
	Token    string
	Channels channels
	Client   client
}{
	Token: "YOUR_TOKEN_HERE",
	Channels: channels{
		Messaging: "1127831380567523408",
	},
	Client: client{
		FPS: 60,
	},
}

type channels struct {
	Messaging string
}
type client struct {
	FPS uint
}
