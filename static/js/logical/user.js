'use strict'


import { CheckAuthUser, GetOnline, GetUser, GetUsers } from '../spa/api.js';
import { FillUsersSection, CreateUserActions } from '../spa/content.js';
import { WriteContent } from '../spa/content.js';
import { ChangeServerName } from './sign.js';
import { Fetching } from '../spa/fetch.js';
import { Route } from '../spa/router.js';
import { InformUser } from '../spa/inform.js';
import { CheckElemContent } from '../postEngine/cursor.js';

const userData = {
    "id": -1,
    "nick": "",
    "firstName": "",
    "lastName": "",
    "age": -1,
    "gender": "",
    "photo": "",
    "carma": 0,
    "role": -1,
    "friends": "",
    "lastActivitie": "",
    "email": "",
    "status": 0,
};

const signBtn = document.querySelector('.sign-btn'),
    signIcon = document.querySelector('.sign-icon'),
    usernameDiv = document.querySelector('.sign-username');

const changeSignSection = () => {
    if (userData.status === 0) {
        usernameDiv.textContent = '';
        signBtn.classList.add('sign-in');
        signBtn.classList.remove('logout');
        signBtn.textContent = "Войти";
        WriteContent(signIcon, `<i class="fa fa-sign-in" aria-hidden="true"></i>`);
    } else {
        usernameDiv.textContent = userData.nick;
        signBtn.classList.remove('sign-in');
        signBtn.classList.add('logout');
        signBtn.textContent = "Выйти";
        WriteContent(signIcon, `<i class="fa fa-sign-out" aria-hidden="true"></i>`);
    }
}

const changeUserData = async(id) => {
    if (id !== undefined) {
        const user = (await GetUser(id))[0];
        for (let [k, v] of Object.entries(user)) userData[k] = v;
        userData.status = 1;
    } else {
        for (let k in userData) userData[k] = '';
        userData.status = 0;
    }
}

function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

function checkNick() {
    return new Promise(resolve => {
        const inter = setInterval(async() => {
            if (userData.nick === "")
                await sleep(100);
            else {
                clearInterval(inter);
                resolve();
            }
        }, 100);
    })
}

export const UserOnline = async(id) => {
    await checkNick();
    const serverName = userData.nick;
    await changeUserData(id);
    changeSignSection();
    CreateUserActions(userData.role);
    await ChangeServerName(serverName, userData.nick, userData.id);
    FillUsersSection(await GetUsers(0, 10, 'wrote'), await GetOnline());
}

export const UserOffline = async() => {
    await changeUserData();
    changeSignSection();
    CreateUserActions();
    FillUsersSection(await GetUsers(0, 10, 'all'), await GetOnline());
}

export const IsUserLogged = async() => {
    const authState = await CheckAuthUser();
    if (authState.state === 1) return await UserOnline(authState.id);
    await UserOffline();
}

export const GetUserData = () => Object.assign({}, userData);
export const GetUserNick = () => userData.nick;
export const GetUserStatus = () => userData.status;
export const SetUserNick = nick => userData.nick = nick;


export const ChangeUserAvatar = e => {
    const loadInput = document.querySelector('.avatar-change-input');
    const imageDiv = e.target;
    const oldphotoDiv = document.querySelector('.input-oldphoto');
    loadInput.click();

    loadInput.addEventListener('change', async(e) => {
        e.stopImmediatePropagation()
        const file = loadInput.files[0];
        if (file) {
            const reader = new FileReader();
            reader.readAsDataURL(file);
            reader.addEventListener('load', e => imageDiv.setAttribute('src', e.target.result));

            const data = new FormData()
            data.append('oldphoto', oldphotoDiv.value);
            data.append('avatar', file)
            const res = await Fetching('/profile/change-avatar', 'POST', data);

            if (res.msg !== "ok") InformUser(res.msg);
            else {
                oldphotoDiv.value = res.link;
                userData.photo = res.link;
            }
        }
    });
}

export const ChangeProfile = async(e) => {
    e.preventDefault();
    const formData = new FormData(e.target);
    for (let v of formData.values()) {
        if (CheckElemContent(v)) {
            InformUser("invalid text");
            return Route('/');
        }
    }
    const res = await Fetching('/profile/change-profile', 'POST', formData);
    if (res.msg !== "ok") InformUser(res.msg);
    else {
        UserOnline(GetUserData().id);
        return Route('/');
    }
}