package router

import (
	"bytes"
	"html/template"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const expectedHome = `
<!DOCTYPE html>
<html>

<head>
    <meta charset="utf-8">
    <title>What is my IP Address ?</title>
</head>

<body>
    <h1>What is my IP ?</h1>
    <hr />
    <h2> Your IPv4 address is: 127.0.0.1</h2>
    <table>
        <tr> <td> Client Port      </td> <td> 1000 </td> </tr>
        <tr> <td> Host             </td> <td> localhost </td> </tr>
    </table>
    <h3> Geolocation </h3>
    <table>
        <tr> <td> Country          </td> <td> A Country </td> </tr>
        <tr> <td> Country Code     </td> <td> XX </td> </tr>
        <tr> <td> City             </td> <td> A City </td> </tr>
        <tr> <td> Latitude         </td> <td> 100 </td> </tr>
        <tr> <td> Longitude        </td> <td> -100 </td> </tr>
        <tr> <td> Postal Code      </td> <td> 00000 </td> </tr>
        <tr> <td> Time Zone        </td> <td> My/Timezone </td> </tr>
    </table>
    <h3> Autonomous System </h3>
    <table>
        <tr> <td> ASN              </td> <td> 0 </td> </tr>
        <tr> <td> ASN Organization </td> <td> My ISP </td> </tr>
    </table>
    <h3> Headers </h3>
    <table>
        <tr> <td> Header1       </td> <td> value1 </td> </tr>
        <tr> <td> Header2       </td> <td> value21 </td> </tr>
        <tr> <td> Header2       </td> <td> value22 </td> </tr>
        <tr> <td> Header3       </td> <td> value3 </td> </tr>
    </table>
</body>
</html>
`

func TestDefaultTemplate(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header = map[string][]string{
		"Header1": {"value1"},
		"Header2": {"value21", "value22"},
		"Header3": {"value3"},
	}

	tmpl, _ := template.New("home").Parse(home)
	response := JSONResponse{
		IP:              "127.0.0.1",
		IPVersion:       4,
		ClientPort:      "1000",
		Country:         "A Country",
		CountryCode:     "XX",
		City:            "A City",
		Latitude:        100,
		Longitude:       -100,
		PostalCode:      "00000",
		TimeZone:        "My/Timezone",
		ASN:             0,
		ASNOrganization: "My ISP",
		Host:            "localhost",
		Headers:         req.Header,
	}

	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, response)

	assert.Nil(t, err)
	assert.Equal(t, expectedHome, buf.String())
}
