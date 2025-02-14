package numbers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nyaruka/phonenumbers"
)

type PhoneNumber string

func (n *PhoneNumber) UnmarshalJSON(b []byte) error {
	numStr := strings.Trim(string(b), `"`)
	num, err := NewPhoneNumber(numStr)
	if err != nil {
		return err
	}
	*n = *num
	return err
}

func (n PhoneNumber) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.String())
}

func (n PhoneNumber) String() string {
	return string(n)
}

func NewPhoneNumber(number string) (*PhoneNumber, error) {
	var num *phonenumbers.PhoneNumber

	number = strings.Replace(number, "+521", "+52", 1)
	number = strings.Replace(number, "521", "+52", 1)
	number = strings.Replace(number, "+549", "+54", 1)
	number = strings.Replace(number, "549", "+54", 1)

	num, err := phonenumbers.Parse(number, "MX")
	if err != nil {
		num, err = phonenumbers.Parse(number, "")
		if err != nil {
			return nil, err
		}
	}

	formmatedNum := phonenumbers.Format(num, phonenumbers.E164)

	if !phonenumbers.IsValidNumber(num) {
		return nil, fmt.Errorf("%s isn't a valid number", formmatedNum)
	}

	// Los numeros de mexico y argentina necesitan un numero extra
	// https://faq.whatsapp.com/1294841057948784?cms_id=1294841057948784&draft=false
	regionCode := phonenumbers.GetRegionCodeForNumber(num)
	switch regionCode {
	case "MX":
		if strings.HasPrefix(formmatedNum, "+52") && !strings.HasPrefix(formmatedNum[3:], "1") {
			formmatedNum = "+521" + formmatedNum[3:]
		}
	case "AR":
		if strings.HasPrefix(formmatedNum, "+54") && !strings.HasPrefix(formmatedNum[3:], "9") {
			formmatedNum = "+549" + formmatedNum[3:]
		}
	}
	n := PhoneNumber(formmatedNum)
	return &n, nil
}
