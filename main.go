package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"

	"main/backend"
	discrypt "main/midend"

	"github.com/nathan-fiscaletti/consolesize-go"
)

const (
	token string = "MTEyNjc0NDg5OTk5NjM1NjYzOA.Gq8B7Q.3LIzrStmAxiBOJJwSZsuETw-YUgzU5uZnIq2mc"
)

var (
	exit bool = false

	nickname      string = "NoNickname"
	encryptionKey string = "ThisKeyIs16Bytes"
	room          string = "DisCrypt"

	termCols, termLines = consolesize.GetConsoleSize()

	chatPane  *widgets.Paragraph = widgets.NewParagraph()
	inputPane *widgets.Paragraph = widgets.NewParagraph()
	debugPane *widgets.Paragraph = widgets.NewParagraph()
	infoPane  *widgets.Paragraph = widgets.NewParagraph()

	uiEvents <-chan ui.Event = ui.PollEvents()

	messagesGot int = 0

	channels = make(map[string]string)
)

func main() {
	channels["rooms"] = "1127831380567523408"
	channels["messaging"] = "1127831380567523408"

	chatPane.Title = "Chat"
	inputPane.Title = "Message"
	debugPane.Title = "Debugging"
	infoPane.Title = "Info"

	if err := ui.Init(); err != nil {
		fmt.Printf("Failed to initialize termui: %v", err)
		os.Exit(11)
	}
	defer ui.Close()

	messages, err := backend.GetMessages(token, channels["messaging"])
	if err != nil {
		chatPane.Text += "Error while getting messages: " + err.Error() + "\n"
		return
	}
	if len(messages) > 0 {
		messagesGot = len(messages) - 1
	}

	go func() {
		for {
			for encryptionKey == "" {
				time.Sleep(time.Millisecond * 10)
			}
			time.Sleep(time.Millisecond * 2500)
			messages, err := backend.GetMessages(token, channels["messaging"])
			if err != nil {
				chatPane.Text += "Error while getting messages: " + err.Error() + "\n"
				return
			}

			for i := len(messages) - messagesGot - 1; i >= 0; i -= 1 {
				message := messages[i]
				content, ok := message["content"].(string)
				if !ok {
					chatPane.Text += "Error while getting message content\n"
					continue
				}

				decrypted, err := backend.Decrypt(content, encryptionKey)
				if err != nil {
					chatPane.Text += "Error while decryptiing text: " + err.Error()
				} else if decrypted, ok := strings.CutPrefix(decrypted, room+"␞"); ok {
					chatPane.Text += decrypted + "\n"
				}
				messagesGot += 1
				var chatPaneHeight int = termLines - 7
				chatLines := strings.Split(chatPane.Text, "\n")
				for {
					if len(chatLines) > chatPaneHeight {
						chatText, ok := strings.CutPrefix(chatPane.Text, chatLines[0]+"\n")
						if !ok {
							chatPane.Text = ""
						} else {
							chatPane.Text = chatText
						}
						chatLines = strings.Split(chatPane.Text, "\n")
					} else {
						break
					}
				}
				debugPane.Text = "Lines:" + fmt.Sprint(len(chatLines)) + "\nHeight:" + fmt.Sprint(chatPaneHeight)
			}

			if messagesGot > len(messages) {
				messagesGot = len(messages)
			}
		}
	}()

	for {
		time.Sleep(time.Microsecond * 16666) // ~60 fps
		go update()
		if exit {
			break
		}
	}
}

func update() {
	termCols, termLines = consolesize.GetConsoleSize()

	chatPane.SetRect(0, 3, termCols-13, termLines-3)
	inputPane.SetRect(0, termLines-3, termCols-13, termLines)
	debugPane.SetRect(termCols-13, 3, termCols, termLines)
	infoPane.SetRect(0, 0, termCols, 3)

	infoPane.Text = fmt.Sprintf("Room:%s | Nickname:%s | Key:%s", room, nickname, encryptionKey)

	go func() {
		e := <-uiEvents
		if e.Type == ui.KeyboardEvent {
			switch e.ID {
			case "<Escape>":
				exit = true
			case "<Backspace>":
				if inputPane.Text != "" {
					inputPane.Text = inputPane.Text[:len(inputPane.Text)-1]
				}
			case "<C-<Backspace>>":
				if inputPane.Text != "" {
					inputPane.Text = inputPane.Text[:len(inputPane.Text)-1]
				}
			case "<Space>":
				inputPane.Text += " "
			case "<Enter>":
				var message string = inputPane.Text
				inputPane.Text = ""
				if message == "/clear" {
					chatPane.Text = ""
				} else if strings.HasPrefix(message, "/room ") {
					room, _ = strings.CutPrefix(message, "/room ")
				} else if strings.HasPrefix(message, "/nickname ") {
					nickname, _ = strings.CutPrefix(message, "/nickname ")
				} else if strings.HasPrefix(message, "/key ") {
					setEncryptionKey, _ := strings.CutPrefix(message, "/key ")
					if len(setEncryptionKey) != 4 && len(setEncryptionKey) != 8 && len(setEncryptionKey) != 16 && len(setEncryptionKey) != 32 {
						inputPane.Text = "Encryption key must be 4, 8, 16, or 32 bytes in size."
					} else {
						encryptionKey = setEncryptionKey
					}
				} else if room != "" {
					if encryptionKey != "" {
						inputPane.Text = ""
						discrypt.SendMessage(token, channels["messaging"], room+"␞"+nickname+": "+message, encryptionKey)
					} else {
						inputPane.Text = "Set an encryption key using /key <encryption key>"
					}
				} else {
					inputPane.Text = "Join a room first! type /room <roomname>"
				}
			default:
				inputPane.Text += e.ID
			}
		}
	}()

	ui.Clear()
	ui.Render(chatPane, inputPane, debugPane, infoPane)
}
