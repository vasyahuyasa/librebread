function librepaymentConfirm(button, payment_id) {
    let url = `/libre/payment/${payment_id}/confirm`;

    librepaymentPost(button, url);
}

function librepaymentCancel(button, payment_id) {
    let url = `/libre/payment/${payment_id}/cancel`;

    librepaymentPost(button, url);
}

function librepaymentSetStatus(button, payment_id, status) {
    let url = `/libre/payment/${payment_id}/status/${status}`;

    librepaymentPost(button, url);
}

function librepaymentPost(button, url) {
    fetch(url, {
        method: 'POST',
    }).then((resp) => {
        if (resp.status === 200) {
            librepaymentHighlightButton(button, 'librepayment-button-ok');
            window.location.reload();
            return;
        }

        librepaymentHighlightButton(button, 'librepayment-button-error');
    }).catch(() => {
        librepaymentHighlightButton(button, 'librepayment-button-error');
    });
}

function librepaymentHighlightButton(button, className) {
    button.classList.remove('librepayment-button-ok', 'librepayment-button-error');
    button.classList.add(className);

    setTimeout(() => {
        button.classList.remove(className);
    }, 1000);
}
