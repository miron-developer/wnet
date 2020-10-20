'use strict'


export const InformUser = text => {
    document.body.insertAdjacentHTML('afterbegin', `<div class="inform">
                                                        <div class="infrom-msg">${text}</div>
                                                    </div>`);

    document.querySelector('.inform').addEventListener('click', () => document.querySelector('.inform').remove());
}