package main

var Config = struct {
	Token    string
	Channels channels
	Client   client
}{
	Token: "MTEyNjc0NDg5OTk5NjM1NjYzOA.Gey5T_.PIVxoLyGdE4HOKvZqRs7O9CuS1rGVa6T460xdk",
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
