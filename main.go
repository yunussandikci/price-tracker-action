package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yunussandikci/fs-database-go/fsdatabase"
)

var (
	fsDatabase fsdatabase.FSDatabase[Database]
	config     *Config
	bot        *tgbotapi.BotAPI
	digitCheck = regexp.MustCompile(`^[0-9]+$`)
)

func main() {
	database, databaseErr := fsDatabase.Read()
	if databaseErr != nil {
		panic(databaseErr)
	}

	if database.Products == nil {
		database.Products = []*Product{}
	}

	handleMessages(&database)

	if writeErr := fsDatabase.Write(database); writeErr != nil {
		panic(writeErr)
	}

	handleCrawl(&database)

	if writeErr := fsDatabase.Write(database); writeErr != nil {
		panic(writeErr)
	}
}

func handleMessages(database *Database) {
	updates, getUpdatesErr := bot.GetUpdates(tgbotapi.NewUpdate(database.BotUpdateOffset + 1))
	if getUpdatesErr != nil {
		panic(getUpdatesErr)
	}

	for _, update := range updates {
		if update.ChannelPost != nil {
			doneMessage := tgbotapi.MessageConfig{
				BaseChat: tgbotapi.BaseChat{
					ChatID:           update.ChannelPost.SenderChat.ID,
					ReplyToMessageID: update.ChannelPost.MessageID,
				},
				Text: config.DoneMessage,
			}

			if update.ChannelPost.Command() == config.AddCommand {
				database.Products = append(database.Products, &Product{
					ChatID: update.ChannelPost.SenderChat.ID,
					URL: strings.TrimSpace(strings.TrimLeft(update.ChannelPost.Text, fmt.Sprintf("/%s",
						update.ChannelPost.Command()))),
				})

				if _, sendErr := bot.Send(doneMessage); sendErr != nil {
					panic(sendErr)
				}
			} else if update.ChannelPost.Command() == config.RemoveCommand {
				for productIdx, product := range database.Products {
					if product.URL == strings.TrimSpace(strings.TrimLeft(update.ChannelPost.Text,
						fmt.Sprintf("/%s", update.ChannelPost.Command()))) {
						database.Products = append(database.Products[:productIdx], database.Products[productIdx+1:]...)

						if _, sendErr := bot.Send(doneMessage); sendErr != nil {
							panic(sendErr)
						}

						break
					}
				}
			}
		}

		database.BotUpdateOffset = update.UpdateID
	}
}

func handleCrawl(database *Database) {
	for _, product := range database.Products {
		if strings.Contains(product.URL, "amazon.com.tr") {
			urlParts := strings.Split(product.URL, "/")
			amazonID := urlParts[len(urlParts)-1]
			url := fmt.Sprintf("https://www.amazon.com.tr/gp/product/ajax?asin=%s&experienceId=aodAjaxMain", amazonID)

			resp, getErr := http.Get(url)
			if getErr != nil {
				panic(getErr)
			}

			if resp.StatusCode != http.StatusOK {
				fmt.Printf("Status is not OK; status:%d", resp.StatusCode)
				return
			}

			doc, err := goquery.NewDocumentFromReader(resp.Body)
			if err != nil {
				log.Fatal(err)
			}

			resp.Body.Close()

			name := strings.TrimSpace(doc.Find("#aod-asin-title-text").Text())
			price := strings.TrimSpace(doc.Find(".a-offscreen").First().Text())

			if product.Price != price && len(price) > 0 && digitCheck.MatchString(string(price[0])) {
				if _, sendErr := bot.Send(tgbotapi.MessageConfig{
					BaseChat: tgbotapi.BaseChat{
						ChatID: product.ChatID,
					},
					Text: fmt.Sprintf("%s\n%s", name, price),
				}); sendErr != nil {
					panic(sendErr)
				}
			}

			product.Price = price
		}
	}

	time.Sleep(config.FetchDelay)
}

func init() {
	config = NewConfig()

	fsDB, fsDBErr := fsdatabase.New[Database](config.DatabaseFile)
	if fsDBErr != nil {
		panic(fsDBErr)
	}

	botAPI, botAPIErr := tgbotapi.NewBotAPI(config.BotToken)
	if botAPIErr != nil {
		panic(botAPIErr)
	}

	fsDatabase = fsDB
	bot = botAPI
}
