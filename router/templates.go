package router

const home = `
<!DOCTYPE html>
<html>

<head>
    <meta charset="utf-8">
    <title>What is my IP Address ?</title>
</head>

<body>
    <h1>What is my IP ?</h1>
    <hr />
    <h2> Your IPv{{ .IPVersion }} address is: {{ .IP }}</h2>
    <table>
        <tr> <td> Client Port      </td> <td> {{ .ClientPort }} </td> </tr>
        <tr> <td> Host             </td> <td> {{ .Host }} </td> </tr>
    </table>
    <h3> Geolocation </h3>
    <table>
        <tr> <td> Country          </td> <td> {{ .Country }} </td> </tr>
        <tr> <td> Country Code     </td> <td> {{ .CountryCode }} </td> </tr>
        <tr> <td> City             </td> <td> {{ .City }} </td> </tr>
        <tr> <td> Latitude         </td> <td> {{ .Latitude }} </td> </tr>
        <tr> <td> Longitude        </td> <td> {{ .Longitude }} </td> </tr>
        <tr> <td> Postal Code      </td> <td> {{ .PostalCode }} </td> </tr>
        <tr> <td> Time Zone        </td> <td> {{ .TimeZone }} </td> </tr>
    </table>
    <h3> Autonomous System </h3>
    <table>
        <tr> <td> ASN              </td> <td> {{ .ASN }} </td> </tr>
        <tr> <td> ASN Organization </td> <td> {{ .ASNOrganization }} </td> </tr>
    </table>
    <h3> Headers </h3>
    <table>
{{- range $key, $value := .Headers }}
    {{- range $content := $value }}
        <tr> <td> {{ $key }}       </td> <td> {{ $content }} </td> </tr>
    {{- end}}
{{- end }}
    </table>
</body>
</html>
`
