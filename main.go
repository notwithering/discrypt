package main

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/nathan-fiscaletti/consolesize-go"
)

func main() {
	// Tui Init
	if err := ui.Init(); err != nil {
		fmt.Println(errorPrefix, err)
		return
	}
	defer ui.Close()

	/* Bot */
	dg, err := discordgo.New("Bot " + Config.Token)
	if err != nil {
		fmt.Println(errorPrefix, err)
		return
	}

	dg.Identify.Intents = discordgo.IntentsGuildMessages
	if err := dg.Open(); err != nil {
		fmt.Println(errorPrefix, err)
		return
	}

	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		content, err := Decrypt(m.Content, status.Key)
		if err != nil {
			if p, ok := panes["input"]; ok {
				p.Text = fmt.Sprintf("%s %s", errorPrefix, err)
				panes["input"] = p
			}
			return
		}

		var obj ConstantObject
		if err := json.Unmarshal([]byte(content), &obj); err != nil {
			if p, ok := panes["input"]; ok {
				p.Text = fmt.Sprintf("%s %s", errorPrefix, err)
				panes["input"] = p
			}
			return
		}
		switch obj.Type {
		case ObjectMessageType:
			var message MessageObject
			if err := json.Unmarshal([]byte(content), &message); err != nil {
				if p, ok := panes["input"]; ok {
					p.Text = fmt.Sprintf("%s %s", errorPrefix, err)
					panes["input"] = p
				}
				return
			}

			if message.Room != status.Room {
				return
			}

			if p, ok := panes["chat"]; ok {
				p.Text = fmt.Sprintf("%s: %s\n", message.Author, message.Content)
				panes["chat"] = p
			}
		}
	})
	/* */

	/* Panes */
	panes["info"] = pane{Pane: widgets.NewParagraph()}
	panes["info"].Pane.Title = "Info"
	panes["chat"] = pane{Pane: widgets.NewParagraph()}
	panes["chat"].Pane.Title = "Chat"
	panes["input"] = pane{Pane: widgets.NewParagraph()}
	panes["input"].Pane.Title = "Message"

	renderOrder = append(renderOrder, panes["chat"].Pane)
	renderOrder = append(renderOrder, panes["info"].Pane)
	renderOrder = append(renderOrder, panes["input"].Pane)
	/* */

	/* Tui loop */
	var exit bool
	go func() {
		var uiEvents <-chan ui.Event = ui.PollEvents()
		for !exit {
			e := <-uiEvents
			if e.Type == ui.KeyboardEvent {
				switch e.ID {
				case "<Space>":
					if p, ok := panes["input"]; ok {
						p.Text += " "
						panes["input"] = p
					}
				case "<Backspace>":
					if len(panes["input"].Text) != 0 {
						if p, ok := panes["input"]; ok {
							p.Text = p.Text[:len(p.Text)-1]
							panes["input"] = p
						}
					}
				case "<Enter>":
					var contentBuffer string = panes["input"].Text
					if p, ok := panes["input"]; ok {
						p.Text = ""
						panes["input"] = p
					}
					contentBuffer = strings.TrimSpace(contentBuffer)

					if strings.HasPrefix(contentBuffer, "/") {
						switch strings.Split(contentBuffer, " ")[0] {
						case "/nick", "/nickname":
							status.Nickname = func() string {
								var n string
								n = strings.TrimPrefix(contentBuffer, "/nickname ")
								if n != contentBuffer {
									return n
								}
								n = strings.TrimPrefix(contentBuffer, "/nick ")
								return n
							}()
						case "/room":
							status.Room = strings.TrimPrefix(contentBuffer, "/room ")
						case "/key":
							status.DisplayKey = strings.TrimPrefix(contentBuffer, "/key ")
							status.Key = DetermineKey(status.DisplayKey)
						}
						continue
					}

					if contentBuffer == "" {
						continue
					}

					buf, err := json.Marshal(MessageObject{
						Room:    status.Room,
						Content: contentBuffer,
						Author:  status.Nickname,
					})
					if err != nil {
						if p, ok := panes["input"]; ok {
							p.Text = fmt.Sprintf("%s %s", errorPrefix, err)
							panes["input"] = p
						}
						continue
					}

					encrypted, err := Encrypt(string(buf), status.Key)
					if err != nil {
						if p, ok := panes["input"]; ok {
							p.Text = fmt.Sprintf("%s %s", errorPrefix, err)
							panes["input"] = p
						}
						continue
					}

					if _, err := dg.ChannelMessageSend(Config.Channels.Messaging, encrypted); err != nil {
						if p, ok := panes["input"]; ok {
							p.Text = fmt.Sprintf("%s %s", errorPrefix, err)
							panes["input"] = p
						}
						continue
					}
				case "<Escape>":
					exit = true
				default:
					if len(e.ID) == 1 {
						if p, ok := panes["input"]; ok {
							p.Text += e.ID
							panes["input"] = p
						}
					}
				}
			}
		}
	}()
	var cycles uint64
	for !exit {
		time.Sleep(time.Second / time.Duration(Config.Client.FPS))

		var width, height int = consolesize.GetConsoleSize()

		var inputHeightOffset int
		if width-21 != 0 {
			inputHeightOffset = len(panes["input"].Text) / (width - 21)
		}

		for _, p := range panes {
			p.Pane.Text = p.Text
		}

		if p, ok := panes["info"]; ok {
			p.Pane.SetRect(0, 0, width, 3)
			p.Text = fmt.Sprintf("Nickname:%s | Room:%s | Key:%s", status.Nickname, status.Room, status.DisplayKey)
			panes["info"] = p
		}
		if p, ok := panes["chat"]; ok {
			p.Pane.SetRect(0, 2, width, height-2)
		}
		if p, ok := panes["input"]; ok {
			p.Pane.SetRect(0, height-3-inputHeightOffset, width, height)
			if int(cycles)%int(Config.Client.FPS) < int(Config.Client.FPS/2) {
				p.Pane.Text = p.Text + "|"
			} else {
				p.Pane.Text = p.Text
			}
			panes["input"] = p
		}

		ui.Clear()
		ui.Render(renderOrder...)

		if cycles < math.MaxUint64 {
			cycles++
		} else {
			cycles = 0
		}
	}
	/* */
}

var (
	panes       = make(map[string]pane)
	renderOrder []ui.Drawable

	status statusStruct
)

type pane struct {
	Text string
	Pane *widgets.Paragraph
}
type statusStruct struct {
	Nickname   string
	Room       string
	DisplayKey string
	Key        string
}

const (
	debugPrefix   string = "\033[0;33m[DEBUG]\033[0m"
	errorPrefix   string = "\033[0;91m[ERROR]\033[0m"
	warningPrefix string = "\033[0;93m[WARNING]\033[0m"
)

func init() {
	status = statusStruct{
		Nickname:   "NoNickname",
		Room:       "DisCrypt",
		DisplayKey: "MyPrivateKey",
		Key:        DetermineKey("MyPrivateKey"),
	}
}
