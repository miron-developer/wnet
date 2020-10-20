'use strict'


import { GetUserNick } from '../logical/user.js';
import { Fetching } from './fetch.js';

const getData = async(action, data) => await Fetching(action, 'GET', data);

export const GetPosts = async(firstIndex = 0, count = 10, tags = 'all') => {
    const data = new URLSearchParams();
    data.append('firstIndex', firstIndex);
    data.append('count', count);
    data.append('tags', tags);
    return await getData('/api/posts', data);
}

export const GetPost = async(id = -1) => {
    const data = new URLSearchParams();
    data.append('postID', id);
    return await getData('/api/post/', data);
}

export const GetComments = async(firstIndex = 0, count = 10, parentId = -1, typeComment = "post") => {
    const data = new URLSearchParams();
    data.append('firstIndex', firstIndex);
    data.append('count', count);
    data.append('parentID', parentId);
    data.append('type', typeComment);
    return await getData('/api/comments', data);
}

export const GetUsers = async(firstIndex = 0, count = 10, criterie = 'all') => {
    const data = new URLSearchParams();
    data.append('firstIndex', firstIndex);
    data.append('count', count);
    data.append('criterie', criterie);
    return await getData('/api/users', data);
}

export const GetOnline = async() => await getData('/api/online', '');

export const GetMessages = async(nickname, firstIndex = 0, count = 10) => {
    if (nickname === undefined) return;
    const data = new URLSearchParams();
    data.append('firstIndex', firstIndex);
    data.append('count', count);
    data.append('nickname', nickname);
    data.append('username', GetUserNick());
    return await getData('/api/messages', data);
}

export const GetUser = async(id = -1) => {
    const data = new URLSearchParams();
    data.append('userID', id);
    return await getData('/api/user/', data);
}

export const CheckAuthUser = async() => await getData('/sign/status', '');