{{ range $key, $value := . }}
<div class="flagResult">
    <h2>{{$value.Country}}</h2>
    <img src="/flags/{{ $value.Image}}.webp">
</div>
{{ end }}