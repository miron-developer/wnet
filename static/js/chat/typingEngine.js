'use strict'


import { SendWSMessage } from "../logical/ws.js";
import { Debounce } from "../spa/content.js";

let status = 0
export const StartTyping = receiver => {
    if (status === 1) return;
    status = 1;
    SendWSMessage(9, receiver, "");
}

export const SendExStopTyping = receiver => {
    status = 0;
    SendWSMessage(10, receiver, "");
}

export const StopTyping = Debounce(receiver => {
    if (status === 0) return;
    status = 0;
    SendWSMessage(10, receiver, "");
}, 2500);