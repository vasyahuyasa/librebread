<table border=1>
    <caption>Telegram Bot API ({{len .}})</caption>
	<thead>
		<th>Time</th>
		<th>Token</th>
        <th>Method</th>
        <th>Payload</th>
	</thead>
    <tbody>
        {{range .}}
        <tr>
            <td>{{.Time}}</td>
            <td>{{.Token}}</td>
            <td>{{.Method}}</td>
            <td><pre>{{.Payload}}<pre></td>
        </tr>
        {{end}}
    </tbody>
</table>