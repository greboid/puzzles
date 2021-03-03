<p>Size: {{$.Width}} x {{$.Height}}<br>
Type: {{$.ImageType}}<br>
Date: {{if .ExifData.Comment}}{{$.ExifData.DateTime}}{{else}}No date/time{{end}}<br>
{{if .ExifData.Comment}}<a href="{{.ExifData.MapLink}}">Map link</a>{{else}}No coordinates present{{end}}<br>
Comment: {{if .ExifData.Comment}}{{$.ExifData.Comment}}{{else}}No comment{{end}}<br>
Raw Values:
<ul>
        {{ range $key, $value := .ExifData.RawValues }}
                <li>{{$key}}: {{$value}}</li>
        {{ end }}
</ul>
</p>