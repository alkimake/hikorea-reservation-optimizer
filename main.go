package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"github.com/joho/godotenv"
)

func checkAvailableDate() (string, error) {

	user_id := os.Getenv("USER_ID")
	password := os.Getenv("PASSWORD")
	desiredMonth, err := strconv.Atoi(os.Getenv("DESIRED_MONTH"))
	if err != nil {
		return "", fmt.Errorf(`error while converting the desired month to integer %s`, err)
	}
	// get the DESIRED_AFTER_DAY_OF_MONTH from env variables or default to 1
	// this is the day of the month after the desired mon
	desiredAfterDayOfMonth := os.Getenv("DESIRED_AFTER_DAY_OF_MONTH")
	if desiredAfterDayOfMonth == "" {
		desiredAfterDayOfMonth = "0"
	}

	minDay, err := strconv.Atoi(desiredAfterDayOfMonth)
	if err != nil {
		return "", fmt.Errorf(`error while converting the desired after day of month to integer %s`, err)
	}

	// create context
	opts := append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Flag("headless", true))
	// silent chromedp logs
	opts = append(opts, chromedp.Flag("enable-logging", false))
	actx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel := chromedp.NewContext(actx)
	defer cancel()

	err = chromedp.Run(ctx,
		chromedp.Navigate(`https://www.hikorea.go.kr/mypage/MypgResvPageR.pt#`),
	)
	if err != nil {
		return "", err
	}

	pop_up_close_button_selector := `#alertClose`
	user_id_selector := `#userId`
	password_selector := `#userPasswd`
	submit_button_selector := `#loginForm > div > div.form_box > div.login_button > a`
	// application_of_sojourn_link := `#content > div.section_content > div > div.grp_content > div > table > tbody > tr > td:nth-child(3) > a`
	edit_button_selector := `#btn_updateResv`
	date_button_selector := `#resvYmdSelect`
	date_input_selector := `#resvYmd`
	logged_in_selector := `#header > div.top > div > div > ul > li:nth-child(1)`
	var dateValue string
	var newWindowTargetID target.ID

	// Get the current list of targets before clicking the button
	initialTargets, err := chromedp.Targets(ctx)
	log.Printf(`initial targets are %v \n`, initialTargets)
	if err != nil {
		return "", err
	}
	err = chromedp.Run(ctx,
		chromedp.WaitVisible(pop_up_close_button_selector, chromedp.ByQuery),
		chromedp.Click(pop_up_close_button_selector, chromedp.ByQuery),
		chromedp.SendKeys(user_id_selector, user_id, chromedp.ByQuery),
		chromedp.SendKeys(password_selector, password, chromedp.ByQuery),
		chromedp.Click(submit_button_selector, chromedp.ByQuery),
		chromedp.WaitVisible(logged_in_selector, chromedp.ByQuery),
		chromedp.Navigate(`https://www.hikorea.go.kr/mypage/MypgResvPageR.pt#`),
		// Wait for the table to be visible
		chromedp.WaitVisible(`.grp_table`, chromedp.ByQuery),
		// Find the last row with Process Status "Reserved" and click the Number link
		chromedp.Click(`//tr[td[contains(text(), '예약')]][last()]/td[1]/a`, chromedp.BySearch),
		// chromedp.WaitVisible(application_of_sojourn_link, chromedp.ByQuery),
		// chromedp.Click(application_of_sojourn_link, chromedp.ByQuery),
		chromedp.WaitVisible(edit_button_selector, chromedp.ByQuery),
		chromedp.Click(edit_button_selector, chromedp.ByQuery),
		chromedp.Value(date_input_selector, &dateValue, chromedp.ByQuery),
		chromedp.WaitVisible(date_button_selector, chromedp.ByQuery),
		chromedp.Click(date_button_selector, chromedp.ByQuery),

		// Wait for a short while for the popup to open
		chromedp.Sleep(2*time.Second),
	)

	if err != nil {
		return "", err
	}

	// Get the new list of targets after the popup opens
	newTargets, err := chromedp.Targets(ctx)
	if err != nil {
		return "", err
	}

	// Find the new popup window by comparing the initial and new targets
	for _, newTarget := range newTargets {
		found := false
		for _, initialTarget := range initialTargets {
			if newTarget.TargetID == initialTarget.TargetID {
				found = true
				break
			}
		}
		if !found && newTarget.Type == "page" {
			newWindowTargetID = newTarget.TargetID
			break
		}
	}

	// Switch context to the new popup window
	if newWindowTargetID != "" {
		popupCtx, cancel := chromedp.NewContext(ctx, chromedp.WithTargetID(newWindowTargetID))
		defer cancel()

		err = chromedp.Run(popupCtx) // Interact with the new popup here
		// ...other actions within the popup window

		if err != nil {
			return "", err
		}

		log.Printf(`new window target id is %s \n`, newWindowTargetID)

		// parse dateValue 2024-11-07 16:00~16:12 to only get the date
		dateValue = dateValue[:10]
		// convert into time
		date, err := time.Parse("2006-01-02", dateValue)
		if err != nil {
			return "", err
		}
		log.Printf(`Reservation Date is %s \n`, date)

		// get the current date and check the month
		currentDate := time.Now()
		log.Printf(`Current Date is %s \n`, currentDate)

		date_picker_month_value_selector := `#datepicker > div > div > div > span.ui-datepicker-month`
		err = chromedp.Run(popupCtx,
			chromedp.WaitVisible(date_picker_month_value_selector, chromedp.ByQuery),
		)
		if err != nil {
			return "", err
		}

		// click previous button until the month is the same as the current month
		var date_picker_month_value string

		// get the date picker month value, parse it to get the digits only and compare it with the current month, until the value is the same with current month press the date picker previous month button
		for {
			err = chromedp.Run(popupCtx,
				chromedp.Text(date_picker_month_value_selector, &date_picker_month_value, chromedp.ByQuery),
			)
			if err != nil {
				return "", err
			}

			// remove the 월 string from the date picker month value and convert to integer
			date_picker_month_value = strings.Replace(date_picker_month_value, "월", "", -1)
			log.Printf(`Date Picker Month Value is %s \n`, date_picker_month_value)
			current_selected_month, err := strconv.Atoi(date_picker_month_value)
			if err != nil {
				return "", fmt.Errorf(`error while converting the date picker month value to integer %s`, err)
			}

			log.Printf(`Date Picker Month Value is %d \n`, current_selected_month)
			log.Printf(`Current Month is %d \n`, desiredMonth)
			if current_selected_month == desiredMonth {
				break
			}
			// sleep for 1 seconds
			time.Sleep(1 * time.Second)
			// click the previous month button
			err = chromedp.Run(popupCtx,
				chromedp.Click(`#datepicker > div > div > a.ui-datepicker-prev.ui-corner-all`, chromedp.ByQuery),
			)
			if err != nil {
				return "", fmt.Errorf(`error while clicking the previous month button %s`, err)
			}
		}

		var firstAvailableDate string

		err = chromedp.Run(popupCtx,
			// Wait for the date picker to be visible
			chromedp.WaitVisible(`.ui-datepicker-calendar`, chromedp.ByQuery),

			// Find the first available date (td with class 'undefined')
			chromedp.EvaluateAsDevTools(`
				(() => {
					const availableDates = document.querySelectorAll('td.undefined a.ui-state-default');
					if (availableDates.length > 0) {
						availableDates[0].click(); // Click the first available date
						return availableDates[0].textContent; // Return the date as a string
					}
					return null;
				})()
			`, &firstAvailableDate),
		)

		// Check if an available date was found and clicked
		if err != nil {
			return "", fmt.Errorf(`error while clicking the first available date %s`, err)
		} else if firstAvailableDate != "" {
			// get only the digit in first available date
			re := regexp.MustCompile("[0-9]+")
			day := re.FindString(firstAvailableDate)
			dayInt, err := strconv.Atoi(day)
			if err != nil {
				return "", fmt.Errorf(`error while converting the day to integer %s`, err)
			}
			if dayInt < minDay {
				return "", fmt.Errorf(`no available dates found`)
			}
			return day, nil
		} else {
			return "", fmt.Errorf(`no available dates found`)
		}

	}

	return "", fmt.Errorf(`no new window target id found`)

}

func check() error {

	day, err := checkAvailableDate()
	if err != nil {
		// SendGenericMessage(err.Error())
		return err
	}
	log.Printf(`Available Date is %s \n`, day)

	// Send the available date to Telegram
	SendToTelegram(day)
	return nil
}

func main() {

	// Load the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	// call check command every 5 minutes
	for {
		err := check()
		if err != nil {
			log.Printf("Error checking available date: %s\n", err)
		}
		time.Sleep(5 * time.Minute)
	}
}
