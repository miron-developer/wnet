'use strict'


import { SetUserNick, GetUserNick } from "./user.js";
import { WriteContent, ContentHeader, CreateMessage, NewUserOnline, NewUserOffline, GenerateTypingAnimation, RemoveTypingAnimation } from "../spa/content.js";
import { InformUser } from "../spa/inform.js";

let wss = null;
let file = '';
let fileInfo = {};

export const CloseWSConnection = () => wss.close();

export const CreateWSConnection = () => {
    const host = window.location.host;
    wss = new WebSocket('wss://' + host + '/ws/');

    wss.onopen = () => console.log("websocket connected");
    wss.onclose = (e) => {
        wss = null;
        WriteContent(ContentHeader, "Websocket");
        WriteContent(undefined, `<h2 class="ws-message">Websocket connection closed!</h2>`);
        console.log('closed', e);
    };

    wss.onerror = err => {
        wss = null;
        WriteContent(ContentHeader, "Websocket");
        WriteContent(undefined, `<h2 class="ws-message">Websocket connection failed! Pls refresh page</h2>`);
        console.log(err);
    };

    wss.onmessage = e => {
        const data = JSON.parse(e.data);

        // handle messages
        if (data.msgType === -1) return InformUser(data.body);
        else if (data.msgType === 0) return SetUserNick(data.body);
        else if (data.msgType === 1) {
            if (window.location.pathname !== '/chat/' + data.addresser) {
                const users = Array.from(document.querySelectorAll('.user'));
                const user = users.find(user => user.childNodes[0].textContent === data.addresser);
                user.childNodes[1].classList.add('is-have-content');
                const curCount = parseInt(user.childNodes[1].textContent) || 0;
                user.childNodes[1].textContent = curCount + 1;
            } else {
                CreateMessage(data.body, data.receiver === GetUserNick() ? true : false);
                return RemoveTypingAnimation();
            }
        } else if (data.msgType === 2) return // TODO: past comments
        else if (data.msgType === 3) return // TODO: handle post/posts
        else if (data.msgType === 4) return NewUserOnline(data.body);
        else if (data.msgType === 5) return NewUserOffline(data.body);
        else if (data.msgType === 6) {
            file = '';
            fileInfo = data.body;
        } else if (data.msgType === 7) file += data.body;
        else if (data.msgType === 8) {
            CreateMessage(`<img src="${file}" />`, data.receiver === GetUserNick() ? true : false)
        } else if (data.msgType === 9) {
            if (window.location.pathname === '/chat/' + data.addresser)
                return GenerateTypingAnimation(document.querySelector('.chat-msgs'), data.addresser);
        } else if (data.msgType === 10) {
            if (window.location.pathname === '/chat/' + data.addresser)
                return RemoveTypingAnimation();
        }
    }
}

export const SendWSMessage = (msgType = 1, receiver = 'all', body) => {
    if (body === undefined || wss === null) return 'dont sended';
    wss.send(JSON.stringify({ "msgType": msgType, "addresser": GetUserNick(), "receiver": receiver, "body": body }));
}