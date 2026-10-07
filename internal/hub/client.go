package hub

type Client struct {
	Name string
	Send chan string
}
