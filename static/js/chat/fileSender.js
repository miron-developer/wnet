'use strict'


import { SendWSMessage } from "../logical/ws.js";
import { CreateMessage } from "../spa/content.js";
import { Fetching } from "../spa/fetch.js";
import { InformUser } from "../spa/inform.js";

export const SendFileOverWS = (file, nick) => {
    if (!file || !nick) return;
    const pocketSize = 64000;
    const pocketCount = file.size / pocketSize;
    if (file) {
        const reader = new FileReader();
        reader.readAsDataURL(file);
        reader.addEventListener('load', e => {
            SendWSMessage(6, nick, { "name": file.name, "type": file.type, "size": file.size, "pocketSize": pocketSize, "pocketCount": pocketCount });
            let data = e.target.result;
            while (data.length >= pocketSize) {
                SendWSMessage(7, nick, data.slice(0, pocketSize));
                data = data.slice(pocketSize);
            }
            SendWSMessage(7, nick, data);
            SendWSMessage(8, nick, { "length": e.target.result.length });
            CreateMessage(`<img src="${e.target.result}" />`);
        });
    }
}

export const SaveAsMsg = async(file, nick) => {
    if (!file || !nick) return;
    let data = new FormData();
    data.append("type", "img");
    data.append("file", file);
    data.append("place", "chatFiles")
    const res = await Fetching('/save/file', 'POST', data)
    if (res.msg !== "ok") return InformUser("wrong to send image");

    data = new FormData();
    data.append("receiver", nick);
    data.append("body", `<img src="/${res.fname}" />`);
    Fetching('/save/message', 'POST', data);
}