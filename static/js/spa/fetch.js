'use strict'


export const Fetching = async(action, method = "POST", data) => {
    const fetchOption = { 'method': method };
    if (method == "GET") action += "?" + data.toString();
    else fetchOption["body"] = data;

    return await fetch(action, fetchOption)
        .then(res => res.json())
        .catch(err => console.log(err, action));
}