<script src="/static/js/librepaymets.js"></script>
<style>
    .librepayment-status-button {
        margin-left: 4px;
        padding: 3px 8px;
        border: 1px solid #8a8f98;
        border-radius: 4px;
        background: #f7f8fa;
        color: #222;
        cursor: pointer;
        font-size: 0.9em;
        line-height: 1.3;
        transition: background 120ms ease, border-color 120ms ease, box-shadow 120ms ease;
    }
    .librepayment-status-dropdown {
        display: inline-block;
        margin-left: 4px;
        position: relative;
    }
    .librepayment-status-dropdown summary {
        padding: 3px 8px;
        border: 1px solid #8a8f98;
        border-radius: 4px;
        background: #f7f8fa;
        cursor: pointer;
        line-height: 1.3;
        transition: background 120ms ease, border-color 120ms ease, box-shadow 120ms ease;
    }
    .librepayment-status-dropdown summary:hover {
        border-color: #4267b2;
        background: #eef3ff;
        box-shadow: 0 0 0 2px rgba(66, 103, 178, 0.16);
    }
    .librepayment-status-menu {
        background: #fff;
        border: 1px solid #8a8f98;
        border-radius: 4px;
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.14);
        display: flex;
        flex-wrap: wrap;
        gap: 4px;
        margin-top: 4px;
        min-width: 320px;
        padding: 6px;
        position: absolute;
        z-index: 10;
    }
    .librepayment-status-button:hover {
        border-color: #4267b2;
        background: #eef3ff;
        box-shadow: 0 0 0 2px rgba(66, 103, 178, 0.16);
    }
    .librepayment-button-ok {
        background: #c8f7c5;
        border-color: #198754;
        color: #0f5132;
    }
    .librepayment-button-error {
        background: #f8d7da;
        border-color: #dc3545;
        color: #842029;
    }
</style>

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
                <button onclick="librepaymentCancel(this, {{.ID}})">Cancel</button>
                <button onclick="librepaymentConfirm(this, {{.ID}})">Confirm</button>
                <details class="librepayment-status-dropdown">
                    <summary>Set status</summary>
                    <div class="librepayment-status-menu">
                        <button class="librepayment-status-button" onclick="librepaymentSetStatus(this, {{.ID}}, 'new')">New</button>
                        <button class="librepayment-status-button" onclick="librepaymentSetStatus(this, {{.ID}}, 'formShowed')">Form</button>
                        <button class="librepayment-status-button" onclick="librepaymentSetStatus(this, {{.ID}}, 'deadlineExpired')">Expired</button>
                        <button class="librepayment-status-button" onclick="librepaymentSetStatus(this, {{.ID}}, 'canceled')">Cancel</button>
                        <button class="librepayment-status-button" onclick="librepaymentSetStatus(this, {{.ID}}, 'authorizing')">Auth</button>
                        <button class="librepayment-status-button" onclick="librepaymentSetStatus(this, {{.ID}}, 'authorized')">Authorized</button>
                        <button class="librepayment-status-button" onclick="librepaymentSetStatus(this, {{.ID}}, 'confirming')">Confirming</button>
                        <button class="librepayment-status-button" onclick="librepaymentSetStatus(this, {{.ID}}, 'rejected')">Rejected</button>
                        <button class="librepayment-status-button" onclick="librepaymentSetStatus(this, {{.ID}}, 'confirmed')">Confirmed</button>
                        <button class="librepayment-status-button" onclick="librepaymentSetStatus(this, {{.ID}}, 'refunding')">Refund</button>
                        <button class="librepayment-status-button" onclick="librepaymentSetStatus(this, {{.ID}}, 'refunded')">Refunded</button>
                    </div>
                </details>
            </td>
        </tr>
    </tbody>
</table>

<table border=1>
    <caption>Status history</caption>
    <thead>
        <tr>
            <th scope="col">Date</th>
            <th scope="col">Status</th>
        </tr>
    </thead>
    <tbody>
        {{ range .StatusHistory}}
        <tr>
            <td>{{.Date}}</td>
            <td>{{.Status}}</td>
        </tr>
        {{end}}
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
