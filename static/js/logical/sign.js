'use strict'



import { Fetching } from '../spa/fetch.js';
import { UserOnline, UserOffline, GetUserData } from './user.js';
import { Route } from '../spa/router.js';
import { InformUser } from '../spa/inform.js';
import { CheckElemContent } from '../postEngine/cursor.js';

export const SignIn = async(e) => {
    e.preventDefault();
    const formData = new FormData(document.querySelector('.sign-in form'));
    for (let v of formData.values()) {
        if (CheckElemContent(v)) {
            InformUser("invalid text");
            return Route('/');
        }
    }

    const res = await Fetching('/sign/in', 'POST', formData);
    if (res.msg != 'ok') InformUser(res.msg);
    else {
        UserOnline(res.id);
        Route('/');
    }
}

const isValideSignUpForm = () => {
    const numberInp = document.querySelector('.sign-up input[type="number"]');
    if (numberInp.value < 0) {
        InformUser("введите правильный возраст(положительный)");
        return false;
    }
    return true;
}

// registration
export const SingUp = async(e) => {
    e.preventDefault();
    if (!isValideSignUpForm()) return;
    const formData = new FormData(document.querySelector('.sign-up form'));

    for (let v of formData.values()) {
        if (CheckElemContent(v)) {
            InformUser("invalid text");
            return Route('/');
        }
    }
    const res = await Fetching('/sign/up', 'POST', formData);
    if (res.msg != 'ok') InformUser(res.msg);
    else Route('/sign/s/');
}

export const SaveUser = async(e) => {
    if (e) e.preventDefault();
    const formData = new FormData(document.querySelector('.save-user form'));

    for (let v of formData.values()) {
        if (CheckElemContent(v)) {
            InformUser("invalid text");
            return Route('/');
        }
    }
    const res = await Fetching('/sign/s/', 'POST', formData);
    if (res.msg != 'ok') InformUser(res.msg);
    else {
        UserOnline(res.id);
        Route('/');
    }
}

// restovration
export const Restore = async(e) => {
    e.preventDefault();
    const formData = new FormData(document.querySelector('.restore form'));

    for (let v of formData.values()) {
        if (CheckElemContent(v)) {
            InformUser("invalid text");
            return Route('/');
        }
    }
    const res = await Fetching('/sign/restore', 'POST', formData);
    if (res.msg != 'ok') InformUser(res.msg);
    else Route('/sign/r/');
}

export const RestorePassword = async(e) => {
    e.preventDefault();
    const formData = new FormData(document.querySelector('.restore-password form'));

    for (let v of formData.values()) {
        if (CheckElemContent(v)) {
            InformUser("invalid text");
            return Route('/');
        }
    }
    const res = await Fetching('/sign/r/', 'POST', formData);
    if (res.msg != 'ok') InformUser(res.msg);
    else {
        InformUser("Ваш пароль изменен! Можете войти через новый пароль");
        // UserOnline(res.id);
        Route('/sign/in');
    }
}

export const Logout = () => {
    const data = new FormData();
    data.append('nickname', GetUserData().nick);
    UserOffline();
    Fetching('/sign/logout', 'POST', data);
}

export const ChangeServerName = async(oldname, newname, id) => {
    const data = new URLSearchParams();
    data.append('oldname', oldname);
    data.append('newname', newname);
    data.append('id', id);
    await Fetching('/ws/change-name', 'POST', data);
}