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
    <caption>LibrePayment ({{len .}})</caption>
	<thead>
		<th>Time</th>
		<th>ID</th>
		<th>Amount</th>
        <th>Merchant</th>
        <th>Status</th>
		<th>Action</th>
	</thead>
    <tbody>
        {{range .}}
        <tr>
            <td>{{.Time}}</td>
            <td><a href="/librepayments/{{.ID}}">{{.ID}}</a></td>
            <td>{{.Amount}}</td>
            <td>{{.Merchant}}</td>
            <td>{{.Status}}</td>
            <td>
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
        {{end}}
    </tbody>
</table>
