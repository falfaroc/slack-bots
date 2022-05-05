package bot

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"pie-bot/pie-bot/src/util"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func calculateDate(date time.Time) time.Time {
	date = date.AddDate(0, 0, 1)

	for date.Weekday() != time.Thursday {
		date = date.AddDate(0, 0, 1)
	}

	return date
}

func updateState(ctx *util.PieContent) {
	if len(ctx.Members) == 0 {
		return
	}

	if len(ctx.Members) == 1 {
		fmt.Println("Get more friends to make pies for you.")
		date := calculateDate(time.Now())
		next_week := date.AddDate(0, 0, 7)

		ctx.State.Current.Date = date.Format("2006-02-01")
		ctx.State.Current.ID = ctx.Members[0].ID

		ctx.State.Next.Date = next_week.Format("2006-02-01")
		ctx.State.Next.ID = ctx.Members[0].ID

		if len(ctx.Pies) != 0 {
			lastPie := ctx.Pies[len(ctx.Pies)-1]

			ctx.State.Previous.Date = lastPie.Date
			ctx.State.Previous.ID = lastPie.Member.ID
		}

		fmt.Println(ctx.State.Current)
		fmt.Println(ctx.State.Next)
		fmt.Println(ctx.State.Previous)

		return
	}

	date := calculateDate(time.Now())
	next_week := date.AddDate(0, 0, 7)

	ctx.State.Current.Date = date.Format("2006-02-01")
	ctx.State.Current.ID = ctx.Members[0].ID

	ctx.State.Next.Date = next_week.Format("2006-02-01")
	ctx.State.Next.ID = ctx.Members[1].ID

	if len(ctx.Pies) != 0 {
		lastPie := ctx.Pies[len(ctx.Pies)-1]

		ctx.State.Previous.Date = lastPie.Date
		ctx.State.Previous.ID = lastPie.Member.ID
	}

	fmt.Println("Let's See!")
	fmt.Println(ctx.State.Current)
	fmt.Println(ctx.State.Next)
	fmt.Println(ctx.State.Previous)
}

func handleAddCommand(id string, name string, command []string, pies *util.PieContent) (error, string) {
	if len(command) < 3 {
		return errors.New("Add command found with invalid length."), ""
	}

	if len(pies.Members) == 0 {
		return errors.New("No one is part of the party."), ""
	}

	if id != pies.State.Current.ID {
		return errors.New("You are not next DAN!"), ""
	}

	date := time.Now()
	pieType := strings.TrimSuffix(strings.Join(command[2:], " "), "\n")
	member := pies.Members[0]

	pies.Pies = append(pies.Pies, util.Pie{Type: pieType, Date: date.Format("2006-02-01"), Member: member})
	pies.Members = append(pies.Members, member)
	pies.Members = pies.Members[1:]

	updateState(pies)

	util.SavePersistentDate(pies)

	response := "What a great " + pieType + " by " + name

	return nil, response
}

func handleJoinCommand(user *slack.User, command []string, pies *util.PieContent) error {
	if len(command) < 2 {
		return errors.New("Join command found with invalid length.")
	}

	for _, val := range pies.Members {
		if user.ID == val.ID {
			return errors.New("You're already part of the club. Get some friends!")
		}
	}

	pies.Members = append(pies.Members, util.Member{ID: user.ID, Name: user.RealName})

	fmt.Println(len(pies.Members))

	updateState(pies)

	util.SavePersistentDate(pies)

	return nil
}

func handleNextCommand(pies *util.PieContent) string {
	if pies.State.Next.ID == "" {
		return "I have no friends. Someone please join the club."
	}

	var name string

	for _, val := range pies.Members {
		if pies.State.Next.ID == val.ID {
			name = val.Name
			break
		}
	}

	return name + " is scheduled for " + pies.State.Next.Date
}

func handleCurrentCommand(pies *util.PieContent) string {
	if pies.State.Current.ID == "" {
		return "I have no friends. Someone please join the club."
	}

	var name string

	for _, val := range pies.Members {
		if pies.State.Current.ID == val.ID {
			name = val.Name
			break
		}
	}

	return name + " is scheduled for " + pies.State.Current.Date
}

func handlePreviousCommand(pies *util.PieContent) string {
	if pies.State.Previous.ID == "" {
		return "No one has made a pies. DISAPPOINTED!"
	}

	var name string

	for _, val := range pies.Members {
		if pies.State.Previous.ID == val.ID {
			name = val.Name
			break
		}
	}

	return "The last pie was made one " + pies.State.Previous.Date + " by " + name
}

func handleHistoryCommand(pies *util.PieContent) string {
	response := fmt.Sprintf("====================One history coming up!====================\n")
	// fmt.Printf("%30s One history coming up! %30s\n", "", "")
	fmt.Printf("%s\n", strings.Repeat("=", 75))

	for _, s := range pies.Pies {
		response += fmt.Sprintf("| %15s | %40s | %20s |\n", s.Date, s.Member.Name, s.Type)
	}

	return response
}

func handleCommand(user *slack.User, command []string, pies *util.PieContent) (slack.Attachment, error) {
	attachment := slack.Attachment{}

	if command[1] == "add" {
		err, response := handleAddCommand(user.ID, user.RealName, command, pies)

		if err != nil {
			attachment.Text = err.Error()
			return attachment, err

		} else {
			attachment.Text = response
			return attachment, nil
		}
	} else if command[1] == "join" {
		err := handleJoinCommand(user, command, pies)

		if err != nil {
			attachment.Text = err.Error()
			return attachment, err

		} else {
			attachment.Text = "New Challenger Approaches:" + user.RealName
			return attachment, nil
		}
	} else if command[1] == "next" {
		attachment.Text = handleNextCommand(pies)
		return attachment, nil
	} else if command[1] == "current" {
		attachment.Text = handleCurrentCommand(pies)
		return attachment, nil
	} else if command[1] == "last" {
		attachment.Text = handlePreviousCommand(pies)
		return attachment, nil
	} else if command[1] == "history" {
		attachment.Text = handleHistoryCommand(pies)
	} else {
		attachment.Text = fmt.Sprintf("SILENCE DAN!")
	}

	return attachment, nil
}

func handleAppMentionEvent(event *slackevents.AppMentionEvent, client *slack.Client, pies *util.PieContent) error {
	user, err := client.GetUserInfo(event.User)
	if err != nil {
		return err
	}

	text := strings.ToLower(event.Text)
	text_slice := strings.Split(text, " ")

	fmt.Printf("\n\nDebug this\nString:%s\nSlice:%d\n\n", text, len(text_slice))

	attachment := slack.Attachment{}

	if len(text_slice) < 2 {
		attachment.Text = fmt.Sprintf("Invalid command.")
		attachment.Color = "#3d3d3d"
	} else {
		attachment, err = handleCommand(user, text_slice, pies)
	}

	_, _, err = client.PostMessage(event.Channel, slack.MsgOptionAttachments(attachment))
	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}
	return nil
}

func handleEventMessage(event slackevents.EventsAPIEvent, client *slack.Client, pies *util.PieContent) error {
	switch event.Type {
	case slackevents.CallbackEvent:

		innerEvent := event.InnerEvent

		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			err := handleAppMentionEvent(ev, client, pies)
			if err != nil {
				return err
			}
		}
	default:
		return errors.New("unsupported event type")
	}

	return nil
}

func Execute() {
	err := godotenv.Load("secret/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	pies := util.LoadPersistentData(util.PersistentFile)
	updateState(&pies)

	token := os.Getenv("SLACK_AUTH_TOKEN")
	appToken := os.Getenv("SLACK_APP_TOKEN")

	fmt.Println(token + " " + appToken)

	client := slack.New(token, slack.OptionDebug(true), slack.OptionAppLevelToken(appToken))
	socketClient := socketmode.New(
		client,
		socketmode.OptionDebug(true),
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func(ctx context.Context, client *slack.Client, socketClient *socketmode.Client) {
		for {
			select {
			case <-ctx.Done():
				log.Println("Shutting down socketmode listener")
				return
			case event := <-socketClient.Events:
				switch event.Type {
				case socketmode.EventTypeEventsAPI:
					eventsAPIEvent, ok := event.Data.(slackevents.EventsAPIEvent)
					if !ok {
						log.Printf("Could not type cast the event to the EventsAPIEvent: %v\n", event)
						continue
					}

					socketClient.Ack(*event.Request)

					err := handleEventMessage(eventsAPIEvent, client, &pies)
					if err != nil {
						log.Fatal(err)
					}
				}

			}
		}
	}(ctx, client, socketClient)

	socketClient.Run()
}
