'use strict'


import { GetComments } from "../spa/api.js";
import { CreateComment, CreateCommentingForm } from "../spa/content.js";
import { Fetching } from "../spa/fetch.js";
import { GetUserStatus } from "./user.js";

let Comments = new Map();
export const StartComments = (id) => {
    Comments = new Map();
    Comments.set(id, 0)
};

export const ShowComments = async(id, whereToPasteComment = '.post-comments .comments', commentType = 'post', setHow = 'afterbegin') => {
    if (id === undefined) return;
    let whatStart = Comments.get(id);
    const comments = await GetComments(whatStart * 10, 11, id, commentType);

    if (comments.length < 11) {
        if (commentType === 'post')
            document.querySelector('.also-comments').classList.add('hidden');
        else
            document.querySelector('.comment-' + id + ' .comment-nest-comment-show').classList.add('hidden');

    }

    Comments.set(id, whatStart + 1);
    if (commentType === 'comment') {
        document.querySelector(whereToPasteComment).insertAdjacentHTML('afterend', `<div class="nested-comments"></div>`);
        whereToPasteComment += ' + .nested-comments';
    }
    comments.slice(0, 10).forEach(comment => {
        const isHaveChild = comment.haveChild
        if (isHaveChild) Comments.set(comment.id, 0)
        CreateComment(comment, whereToPasteComment, setHow);
    });
}

export const MakeCommenting = async(id, wherePastForm = '.your-comment', whereToPasteComment = ".post-comments .comments") => {
    const status = GetUserStatus();
    if (status === 0 || id === undefined) return;
    CreateCommentingForm(id, wherePastForm, whereToPasteComment);
}

const handleLD = async(path, ldtype) => {
    const elem = path.find(elem => /post-\d/.test(elem.className) || /comment-\d/.test(elem.className) || /user-\d/.test(elem.className));
    if (elem === undefined) return;

    const typeAndID = elem.className.split('-');
    const data = new FormData();
    data.append('id', typeAndID[1]);
    data.append('type', typeAndID[0]);
    data.append('value', ldtype);
    const res = await Fetching('/save/ld', 'POST', data);
    if (res.msg !== 'ok') return;

    document.querySelector(`.${typeAndID[0]}-${typeAndID[1]} .${typeAndID[0]}-carma-value`).textContent = res.carma;
}

const getPath = e => e.path || (e.composedPath && e.composedPath());
export const HandleLike = async(e) => handleLD(getPath(e), 'like');
export const HandleDislike = async(e) => handleLD(getPath(e), 'dislike');