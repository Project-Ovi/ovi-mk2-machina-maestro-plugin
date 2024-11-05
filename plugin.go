package ovimk2machinamaestroplugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/Project-Ovi/Machina-Maestro/helper"
	"golang.org/x/image/colornames"
)

var thisModel *helper.Model
var workingDirectory string
var defaultCommands *[]helper.Command

func Initi(wd string, model *helper.Model) {
	thisModel = model
	workingDirectory = wd
}

func Form(form *fyne.Container) {
	// Add IP entry
	ipName := canvas.NewText("IP", theme.Color(theme.ColorNameForeground))
	ipEntry := widget.NewEntry()
	ipEntry.Text = "192.168.4.1"
	ipEntry.Validator = validation.NewRegexp("^((https?|ftp):\\/\\/)?((([a-zA-Z0-9-]+\\.)+[a-zA-Z]{2,6})|(localhost)|(\\b((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\b)|(\\b(([0-9a-fA-F]{1,4}:){7}([0-9a-fA-F]{1,4}|:)|(::([0-9a-fA-F]{1,4}:){0,6}([0-9a-fA-F]{1,4}|:)))\\b))(:(\\d{1,5}))?(\\/[a-zA-Z0-9_.~%-]*)*(\\?[a-zA-Z0-9_.~%-&=]*)?(\\#[a-zA-Z0-9_-]*)?$", "Invalid IP/URL")
	ipEntry.Refresh()

	form.Objects = append(form.Objects,
		ipName, ipEntry,
	)
}

func Save(form *fyne.Container) string {
	//
	canContinue := true
	errorMsg := ""

	// Get model name and check if it is unique
	name := form.Objects[1].(*widget.Entry).Text
	if name == "" {
		name = "Unnamed"
	}
	models, err := os.ReadDir("myModels")
	if err != nil {
		log.Fatal(err)
	}
	for _, val := range models {
		if val.IsDir() && val.Name() == name {
			// The name is not unique
			form.Objects[0].(*canvas.Text).Color = colornames.Red
			canContinue = false
			errorMsg = "A model with this name already exists"
		}
	}
	if canContinue {
		form.Objects[0].(*canvas.Text).Color = theme.Color(theme.ColorNameForeground)
	}

	// Get model IP and check it
	IP := form.Objects[7].(*widget.Entry).Text
	form.Refresh()
	if form.Objects[7].(*widget.Entry).Validate() != nil {
		form.Objects[6].(*canvas.Text).Color = colornames.Red
		canContinue = false
		errorMsg = "Invalid IP address"
	} else {
		form.Objects[6].(*canvas.Text).Color = theme.Color(theme.ColorNameForeground)
	}

	// If something is wrong, quit
	if !canContinue {
		return errorMsg
	}

	// Parse model information
	(*thisModel).Name = name
	(*thisModel).Model = "OVI MK2"
	(*thisModel).Website = "https://raw.githubusercontent.com/Project-Ovi/OVI-MK2/refs/heads/main/README.md"
	(*thisModel).Other = map[string]string{"IP": IP}

	// Save model to file
	os.Mkdir(path.Join(workingDirectory, "/myModels/", name), os.ModePerm)
	information, err := json.Marshal(*thisModel)
	if err != nil {
		errorMsg = err.Error()
	}
	err = os.WriteFile(path.Join(workingDirectory, "/myModels/", name, "/model.json"), information, os.ModePerm)
	if err != nil {
		errorMsg = err.Error()
	}
	err = os.WriteFile(path.Join(workingDirectory, "/myModels/", name, "/actions.json"), []byte("[{}]"), os.ModePerm)
	if err != nil {
		errorMsg = err.Error()
	}

	// Return any error messages
	return errorMsg
}
func Load() {
	// * Helper functions
	post := func(url string, args map[string]string) error {
		payload := []byte(`{"key1":"value1", "key2":"value2"}`)

		// Create a new POST request
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
		if err != nil {
			return fmt.Errorf("error creating request: %w", err)
		}

		// Add custom headers from the args map
		for key, value := range args {
			req.Header.Set(key, value)
		}

		// Create an HTTP client
		client := &http.Client{}

		// Send the request
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("error sending request: %w", err)
		}
		defer resp.Body.Close()

		return nil
	}

	// * Load commands
	// Load rotate command
	rotCom := helper.Command{
		DisplayName: "Set rotation speed",
		Arguments: []helper.Argument{
			{Name: "Speed", ArgType: "int", Value: int(0)},
		},
		F: func(a []helper.Argument) error {
			// Fix up data
			if val, ok := a[0].Value.(float64); ok {
				a[0].Value = int(val)
			}

			var R1, R2 int
			if a[0].Value.(int) > 0 {
				R1 = a[0].Value.(int)
				R2 = 0
			} else {
				R1 = 0
				R2 = -a[0].Value.(int)
			}

			if R1 > 255 || R2 > 255 {
				return fmt.Errorf("value out of range. allowed: -256 < speed < 256")
			}

			return post((*thisModel).Other["IP"], map[string]string{
				"R1": fmt.Sprint(R1),
				"R2": fmt.Sprint(R2),
			})
		},
	}

	// Load move up command
	moveupCom := helper.Command{
		DisplayName: "Move up with speed",
		Arguments: []helper.Argument{
			{Name: "Speed", ArgType: "int", Value: int(0)},
		},
		F: func(a []helper.Argument) error {
			// Fix up data
			if val, ok := a[0].Value.(float64); ok {
				a[0].Value = int(val)
			}

			var U1, U2 int
			if a[0].Value.(int) > 0 {
				U1 = a[0].Value.(int)
				U2 = 0
			} else {
				U1 = 0
				U2 = -a[0].Value.(int)
			}

			if U1 > 255 || U2 > 255 {
				return fmt.Errorf("value out of range. allowed: -256 < speed < 256")
			}

			return post((*thisModel).Other["IP"], map[string]string{
				"U1": fmt.Sprint(U1),
				"U2": fmt.Sprint(U2),
			})
		},
	}

	// Load extend forward command
	extendCom := helper.Command{
		DisplayName: "Extend forward with speed",
		Arguments: []helper.Argument{
			{Name: "Speed", ArgType: "int", Value: int(0)},
		},
		F: func(a []helper.Argument) error {
			// Fix up data
			if val, ok := a[0].Value.(float64); ok {
				a[0].Value = int(val)
			}

			var E1, E2 int
			if a[0].Value.(int) > 0 {
				E1 = a[0].Value.(int)
				E2 = 0
			} else {
				E1 = 0
				E2 = -a[0].Value.(int)
			}

			if E1 > 255 || E2 > 255 {
				return fmt.Errorf("value out of range. allowed: -256 < speed < 256")
			}

			return post((*thisModel).Other["IP"], map[string]string{
				"E1": fmt.Sprint(E1),
				"E2": fmt.Sprint(E2),
			})
		},
	}

	// Load grip command
	gripCom := helper.Command{
		DisplayName: "Set gripper state",
		Arguments: []helper.Argument{
			{Name: "Grip", ArgType: "bool", Value: bool(false)},
		},
		F: func(a []helper.Argument) error {
			var G1 = int(0)

			if a[0].Value.(bool) {
				G1 = int(255)
			}

			return post((*thisModel).Other["IP"], map[string]string{
				"G1": fmt.Sprint(G1),
			})
		},
	}

	// Load wait time command
	waitCom := helper.Command{
		DisplayName: "Wait",
		Arguments: []helper.Argument{
			{Name: "Milliseconds", ArgType: "int", Value: int(0)},
		},
		F: func(a []helper.Argument) error {
			// Fix up data
			if val, ok := a[0].Value.(float64); ok {
				a[0].Value = int(val)
			}

			time.Sleep(time.Millisecond * time.Duration(a[0].Value.(int)))
			return nil
		},
	}

	// Append Commands to the default commands list
	(*defaultCommands) = append((*defaultCommands), rotCom, moveupCom, extendCom, gripCom, waitCom)
}
