'use strict'


import { ShowComments, MakeCommenting, HandleLike, HandleDislike } from "../logical/post.js";
import { GetUserData, GetUserNick, GetUserStatus } from "../logical/user.js";
import { GetMessages, GetPost, GetPosts } from "./api.js";
import { Fetching } from "./fetch.js";
import { InformUser } from "./inform.js";
import { Route } from "./router.js";

export const ContentHeader = document.querySelector('.content-header');
export const ContentBody = document.querySelector('.content-body');

export const CleanContent = whatClear => whatClear.innerHTML = '';
export const WriteContent = (where = ContentBody, insertHtml = '') => {
    CleanContent(where);
    where.insertAdjacentHTML('beforeend', insertHtml);
}

export const RemovePreloader = () => document.querySelector('.preloader').remove();
export const GeneratePreloader = whereToPlace => {
    whereToPlace.insertAdjacentHTML('beforeend', `<div class="preloader">
                                                    <div class="circle circle-1"></div>
                                                    <div class="circle circle-2"></div>
                                                    <div class="circle circle-3"></div>
                                                    <div class="circle circle-4"></div>
                                                    <div class="circle circle-5"></div>
                                                    <div class="circle circle-6"></div>
                                                    <div class="circle circle-7"></div>
                                                    <div class="circle circle-8"></div>
                                                </div>`);
}

export const OnlineUsers = {
    onlines: []
};

export const FillUsersSection = async(users = [], onlines = []) => {
    const nickname = GetUserNick();
    let content = '';
    OnlineUsers.onlines = onlines

    users.forEach(
        user => {
            if (user.nick !== nickname) {
                let isOnline = onlines.find(on => on.Nickname === user.nick) !== undefined ? true : false
                content += `<div class="user user-${user.id} ${isOnline?'online':''}"><span class="user-nickname">${user.nick}</span><span class="user-notification"></span></div>`;
            }
        });
    WriteContent(document.querySelector('.chat-users-users'), content);

    if (GetUserStatus())
        document.querySelectorAll('.chat-users-users').forEach(
            user => user.addEventListener('click', e => {
                const elem = e.path.find((elem) => /user-\d/.test(elem.className));
                if (elem !== undefined)
                    return Route('/user/' + elem.classList[1].split("-")[1]);
            })
        );
}

export const CreateUserActions = async(role = -1) => {
    let content = '';
    if (role === -1) content += '<div class="account-actions noaction">Вы не вошли в систему!</div>';

    // mb moders actions
    if (role === 0)
        content += `<div class="account-actions action-admin">Зайти в админскую панель</div>`;

    if (role <= 2 && role >= 0)
        content += `<div class="account-actions action-change-profile">Изменить данные аккаунта</div>
                    <div class="account-actions action-create-post">Создать пост</div>
                    <div class="account-actions action-my-posts">Мои посты</div>`;

    WriteContent(document.querySelector('.navs-account-actions'), content);

    const profBtn = document.querySelector('.action-change-profile');
    if (profBtn) profBtn.addEventListener('click', () => Route('/profile/change-profile'));

    const createPostBtn = document.querySelector('.action-create-post');
    if (createPostBtn) createPostBtn.addEventListener('click', () => Route('/create-post'));

    const myPostsBtn = document.querySelector('.action-my-posts');
    if (myPostsBtn) myPostsBtn.addEventListener('click', () => Route('/profile/my-posts'));
}

export const GenTags = (tags = '') => tags.split(',').reduce((res, tag) => res += `<span class='tag tag-${tag}'>${tag}</span>,`, '');
export const GeneratePosts = posts => {
    const postsContent = posts.reduce((content, post) =>
        content += `<div class="post post-${post.id}">
                        <h2 class="post-title">${post.title}</h2>
                        <div class="post-date">${post.date}</div>
                        <div class="post-tags-wrapper">
                            <h3 class="post-tags-header">Tags:</h3>
                            <div class="post-tags">${GenTags(post.tags).slice(0, -1)}</div>
                        </div>
                        <div class="post-carma-wrapper">
                            <i class="fa fa-heart" aria-hidden="true"></i>
                            <span class="post-carma-value">${post.carma}</span>
                        </div>
                    </div>`, '');
    return postsContent;
}

export const Tags = new Map(Object.entries({
    'Аниме': 'anime',
    'Манга': 'manga',
    'Манхуа': 'manhua',
    'Манхва': 'manhva',
    'Додзинси': 'dodzinci',
    'Комиксы': 'comics',
    'Ранобе': 'ranobe',
    'Новелла': 'lite-novel',
    'ТВ': 'tv',
    'ОВА': 'ova',
    'Вампиры': 'vampire',
    'Военное': 'military',
    'Гарем': 'harem',
    'Детское': 'baby',
    'Гендерная интрига': 'gender-bender',
    'Драма': 'drama',
    'Комедия': 'comedy',
    'Магия': 'magic',
    'Машины': 'machines',
    'Меха': 'mecha',
    'Повседневность': 'dailyLife',
    'Приключение': 'adventure',
    'Романтика': 'romance',
    'Седзё': 'shojo',
    'Седзё Ай': 'shojoAi',
    'Сенен': 'shonen',
    'Сейнен': 'seinen',
    'Спорт': 'sport',
    'Триллер': 'triller',
    'Ужасы': 'horror',
    'Фантастика': 'fantastic',
    'Экшн': 'action',
    'Школа': 'school',
    'Этти': 'etti',
    'Юри': 'yuri',
    'Яой': 'yaoy',
}));

export const GenerateDatalistTag = async(index = 1) => {
    let dl = '';
    Tags.forEach((engl, rus) => dl += `<option value="${engl}">${rus}</option>`)
    document.querySelector('.generate-new-tag').insertAdjacentHTML("beforebegin", ` <label>
                                                                                        <span>Tag</span>
                                                                                        <input type="text" name="tag" list="create-post-tag-${index}" class="created-tag tag-${index}"></input>
                                                                                        <datalist id="create-post-tag-${index}">${dl}</datalist>
                                                                                    </label>`);
}

export const CreatePagination = () => ` <div class="pagination">
                                            <div class="page-left hidden">&#8592;</div>
                                            <div class="page-number">1</div>
                                            <div class="page-right">&#8594;</div>
                                        </div>`;

export const HandlePosts = async() => {
    document.querySelectorAll('.posts > *').forEach(
        post => post.addEventListener(
            'click',
            () => Route('/post/' + post.classList[1].split('-')[1])
        )
    );
}

export const HandlePrevPosts = (firstIndex, where, posts) => {
    if (!where || !posts) return;
    WriteContent(where, GeneratePosts(posts));
    document.querySelector('.page-number').textContent = firstIndex;
    HandlePosts();
}

export const HandleNextPosts = async(firstIndex, where, tags = "all") => {
    const posts = await GetPosts(firstIndex * 10, 11, tags);
    WriteContent(where, GeneratePosts(posts));
    document.querySelector('.page-number').textContent = firstIndex + 1;
    HandlePosts();
    return posts;
}

// commenting
export const CreateComment = (comment, whereToPasteComment = '.post-comments .comments', setHow = 'afterbegin') => {
    if (comment === undefined) return;
    const isHaveChild = comment.haveChild

    document.querySelector(whereToPasteComment).insertAdjacentHTML(
        setHow,
        `<div class="comment-${comment.id}">
            <div class="user-avatar">
                <img src="/${comment.userPhoto}" alt="userPhoto">
                <span>${comment.nickname}</span>
            </div>

            <div class="comment-content">
                <div class="comment-body">${comment.body}</div>

                <div class="comment-bottom">
                    <div class="comment-date">Дата: ${comment.date}</div>

                    <div class="comment-carma">
                        <div class="comment-carma-wrapper">
                            <i class="fa fa-heart" aria-hidden="true"></i>
                            <span class="comment-carma-value">${comment.carma}</span>
                        </div>
                        <div class="comment-actions">
                            <div class="comment-carma-actions">
                                <div class="comment-like">
                                    <i class="fa fa-thumbs-up" aria-hidden="true"></i>
                                </div>
                                <div class="comment-dislike">
                                    <i class="fa fa-thumbs-down" aria-hidden="true"></i>
                                </div>
                            </div>

                            <div class="comment-nest-commenting"><h2>Комментировать?</h2></div>
                            ${isHaveChild === 1 ? '<div class="comment-nest-comment-show"><h2>Показать вложенные коментарии?</h2></div>' : ""}
                        </div>
                    </div>
                </div>
            </div>
        </div>`
    );

    document.querySelector(`.comment-${comment.id} .comment-nest-commenting`).addEventListener(
        'click',
        () => MakeCommenting(comment.id, `.comment-${comment.id}`, `.comment-${comment.id}`)
    );

    if (isHaveChild)
        document.querySelector(`.comment-${comment.id} .comment-nest-comment-show`).addEventListener(
            'click',
            () => {
                if (document.querySelector(`.comment-${comment.id} + .nested-comments`) === null)
                    ShowComments(comment.id, `.comment-${comment.id}`, "comment");
            }
        );

    document.querySelector(`.comment-${comment.id} .comment-like`).addEventListener('click', HandleLike);
    document.querySelector(`.comment-${comment.id} .comment-dislike`).addEventListener('click', HandleDislike);
}

export const CreateCommentingForm = (id, wherePastForm = '.your-comment', whereToPasteComment = ".post-comments .comments") => {
    document.querySelector(wherePastForm).insertAdjacentHTML(
        'afterend',
        `<form class="form">
            <textarea name="body" placeholder="your comment here" required></textarea>
            <input type="submit" value="Отправить">
        </form>`
    );

    const handleCreateCommentForm = async(e) => {
        e.preventDefault();
        if (e.target === undefined) return;
        let commentType = "post"
        if (wherePastForm !== '.your-comment') commentType = "comment"

        const data = new FormData(e.target);
        data.append("id", id);
        data.append('type', commentType);

        const res = await Fetching('/save/comment', 'POST', data);
        if (res.msg != "ok") return InformUser(res.msg);
        if (commentType !== "post") e.target.remove();
        else e.target.reset();

        if (commentType === 'comment') {
            const whereToPasteCommentElement = document.querySelector(whereToPasteComment);
            whereToPasteComment += ' + .nested-comments';
            if (document.querySelector(whereToPasteComment) === null)
                whereToPasteCommentElement.insertAdjacentHTML('afterend', `<div class="nested-comments"></div>`);

        }

        const userdata = GetUserData()
        res.comment["userPhoto"] = userdata.photo
        res.comment["nickname"] = userdata.nick
        CreateComment(res.comment, whereToPasteComment);
    }

    document.querySelector(wherePastForm + '+ .form').addEventListener('submit', handleCreateCommentForm);
}

const isTwoDigit = data => data < 10 ? "0".concat(data) : data

const FormattedDateFromDate = date => "".concat(
    date.getFullYear(), "-", isTwoDigit(date.getMonth() + 1), "-", isTwoDigit(date.getDate()),
    "\n", isTwoDigit(date.getHours()), "-", isTwoDigit(date.getMinutes()), "-", isTwoDigit(date.getSeconds())
);

export const CreateMessage = (msg, reverse = false) => {
    const date = new Date();
    const chat = document.querySelector('.chat-msgs');
    chat.insertAdjacentHTML('beforeend', `<div class="chat-msg ${reverse?'reverse': ''}">
                                            <span class ="msg-date">${FormattedDateFromDate(date)}</span>    
                                            <span class ="msg-msg">${msg}</span>
                                        </div>`);
    chat.scrollTo(0, chat.scrollHeight);
}

export const NewUserOnline = newUser => {
    const users = Array.from(document.querySelectorAll('.user'));
    OnlineUsers.onlines.push(newUser);

    let isUserExist;
    for (let user of users)
        if (user.classList[1].split('-')[1] == newUser.ID) isUserExist = user;

    if (isUserExist === undefined)
        return document.querySelector('.chat-users-users').insertAdjacentHTML(
            'beforeend',
            `<div class="user user-${newUser.ID} online">${newUser.Nickname}</div>`
        );
    isUserExist.classList.add('online');
}

export const NewUserOffline = newUser => {
    const users = Array.from(document.querySelectorAll('.user'));

    let index = -1;
    OnlineUsers.onlines.forEach((user, i) => user.ID == newUser.ID ? index = i : null)

    if (index >= 0)
        OnlineUsers.onlines.splice(index, 1)


    for (let user of users)
        if (user.classList[1].split('-')[1] == newUser.ID)
            user.classList.remove('online');
}

export function Debounce(fn, time) {
    let timeOut;
    return async(...args) => {
        clearTimeout(timeOut);
        timeOut = setTimeout(async() => { await fn(...args) }, time - 1)
    }
}

let stopLoad = false;
export const ResetLoad = () => stopLoad = false;
export const LazyLoad = Debounce(async(nickname, ID, chat, firstIndex, count = 10) => {
    if (stopLoad) return;
    let msgs = await GetMessages(nickname, firstIndex, count);
    if (!msgs) return;
    msgs = msgs.reverse();
    const msgsDiv = msgs.reduce((acc, msg) => acc += `<div class="chat-msg ${msg.receiver===ID?'reverse':''}">
                                                        <span class="msg-date">${msg.date}</span>
                                                        <span class="msg-msg">${msg.body}</span>
                                                    </div>`, '');
    chat.insertAdjacentHTML('afterbegin', msgsDiv);
    if (msgs.length < 10) stopLoad = true;
}, 100);

export const GenerateTypingAnimation = (whereToPlace, username) => {
    if (!whereToPlace || !username) return;
    whereToPlace.insertAdjacentHTML('beforeend', `  <div class="typing-in-progress">
                                                        <div class="typing-wrapper">
                                                            <div class="typing-dot dot-1"></div>
                                                            <div class="typing-dot dot-2"></div>     
                                                            <div class="typing-dot dot-3"></div>
                                                        </div>
                                                        <div class="typing-user-name">${username}</div>
                                                    </div>`);

    whereToPlace.scrollTo(0, whereToPlace.scrollHeight);
}
export const RemoveTypingAnimation = () => {
    const typingAnim = document.querySelector('.typing-in-progress');
    if (typingAnim) typingAnim.remove();
}