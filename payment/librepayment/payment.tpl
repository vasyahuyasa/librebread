<script src="/static/js/librepaymets.js"></script>

<table border=1>
    <caption>{{.ID}}</caption>
    <tbody>        
        <tr>
            <td><b>Time</b></td>
            <td>{{.Time}}</td>
        </tr>
        <tr>
            <td><b>Amount</b></td>
            <td>{{.Amount}}</td>
        </tr>
        <tr>
            <td> <b>Merchant</b></td>
            <td>{{.Merchant}}</td>
        </tr>
        <tr>
            <td><b>Status</b></td>
            <td>{{.Status}}</td>
        </tr>
        {{ range $key, $value := .Payload}}
        <tr>
            <td>{{$key}}</td>
            <td>{{$value}}</td>
        </tr>
        {{end}}
        <tr>
            <td colspan="2">
                <button onclick="librepaymentReject({{.ID}})">Reject</button>
                <button onclick="librepaymentConfirm({{.ID}})">Confirm</button>
            </td>
        </tr>
    </tbody>
</table>

<table border=1>
    <caption>Notification journal</caption>
    <thead>
        <tr>
            <th scope="col">TriedAt</th>
            <th scope="col">TryNum</th>
            <th scope="col">Code</th>
            <th scope="col">Response</th>
            <th scope="col">Error</th>
        </tr>
    </thead>
    <tbody>
        {{ range .Journal}}
        <tr>
            <td>{{.TriedAt}}</td>
            <td>{{.TryNum}}</td>
            <td>{{.Code}}</td>
            <td>{{.Response}}</td>
            <td>{{.Error}}</td>
        </tr>
        {{end}}
    </tbody>
</table>
