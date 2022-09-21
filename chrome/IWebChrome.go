package chrome

import "github.com/tebeka/selenium"

type IWebDriver interface {
	Open(url string) error
	ClickXPath(path string)
	GetXPathText(path string) (text string, err error)
	ViewTips() ([]string, string)
	RadioGetOptions() ([]string, error)
	FillInBlank(answers []string)
	SpecialFillInBlank(answers []string)
	RadioCheck(radioTips string)

	selenium.WebDriver
}
