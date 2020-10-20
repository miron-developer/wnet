'use strict'


import { GetMessages, GetPost, GetPosts, GetUser } from './api.js';
import { Route } from './router.js';
import { SingUp, SignIn, Restore, SaveUser, RestorePassword } from '../logical/sign.js';
import { RemovePreloader, WriteContent, ContentHeader, GeneratePosts, GenerateDatalistTag, HandlePosts, GenTags, CreateMessage, OnlineUsers, LazyLoad, ResetLoad, HandlePrevPosts, HandleNextPosts, CreatePagination, GenerateTypingAnimation } from './content.js';
import { GetUserStatus, GetUserData, ChangeUserAvatar, ChangeProfile } from '../logical/user.js';
import { InitPE } from '../postEngine/engine.js';
import { SendWSMessage } from '../logical/ws.js';
import { ShowComments, MakeCommenting, HandleLike, HandleDislike, StartComments } from '../logical/post.js';
import { InformUser } from './inform.js';
import { SendExStopTyping, StartTyping, StopTyping } from '../chat/typingEngine.js';
import { SaveAsMsg, SendFileOverWS } from '../chat/fileSender.js';


const addCSS = (...links) => {
    const haveLinks = Array.from(document.querySelectorAll('link'));
    links.forEach(flink => {
        const isHave = haveLinks.find(link => link.href.split('/')[5] === flink + ".css");
        if (!isHave) document.head.insertAdjacentHTML('beforeend', `<link rel="stylesheet" href="/static/css/${flink}.css">`);
    });
}
const removeCSS = (...links) => {
    const haveLinks = Array.from(document.querySelectorAll('link'));
    links.forEach(flink => {
        const isHave = haveLinks.find(link => link.href.split('/')[5] === flink + ".css");
        if (isHave) isHave.remove();
    });
}

// page generators
export const NotFoundPage = async() => {
    WriteContent(ContentHeader, '404');
    WriteContent(undefined, '<h2 class="NF-h2"> Omaeva mou shindeiru </h2>');
    RemovePreloader();
}

let topPosts = [];
let gettedTop = false;
export const MainPage = async() => {
    addCSS('index', 'posts', 'typing');
    WriteContent(ContentHeader, 'Главная');
    WriteContent(undefined, `<div class="about-anifor">
                                <h2 class="about-anifor-h2">Про форум AniFor</h2>
                                <p> Как и предыдущий <a href="https://miron-forum.herokuapp.com/" target="_blank">форум</a>, сделан на тематику японского творчества, а именно: аниме, манга и тд.</p>
                                <p> Форум сделан согласно заданию и навыками программистов</p>
                            </div>
                            <div class="top-posts">
                                <h2> Топ 10 постов по релевантности </h2>
                                <div class="posts"></div>
                            </div>`);

    if (!gettedTop) {
        topPosts = await GetPosts(0, 10, 'top');
        gettedTop = true;
    }
    RemovePreloader();
    if (topPosts.length === 0) return WriteContent(document.querySelector('.top-posts'), 'Постов нет')
    WriteContent(document.querySelector('.posts'), GeneratePosts(topPosts));
    HandlePosts();
}

let tenPosts = [];
let gettedTen = false;
export const PostsPage = async() => {
    let currentTPPage = 1;
    addCSS('posts');
    WriteContent(ContentHeader, 'Посты')
    WriteContent(undefined, `   <div class="posts"></div>
                                ${CreatePagination()}`);
    if (!gettedTen) {
        tenPosts = await GetPosts(0, 11, 'all');
        gettedTen = true;
    }
    RemovePreloader();
    if (tenPosts.length === 0) return WriteContent(document.querySelector('.content-body'), 'Постов нет');
    if (tenPosts.length < 11) document.querySelector('.page-right').classList.add("hidden");
    WriteContent(document.querySelector('.posts'), GeneratePosts(tenPosts.slice(0, 10)));
    HandlePosts();

    // handle pagination
    document.querySelector('.page-left').addEventListener('click', e => {
        currentTPPage--;
        if (currentTPPage > 0) {
            HandlePrevPosts(currentTPPage, document.querySelector('.posts'), tenPosts.slice((currentTPPage - 1) * 10, currentTPPage * 10));
            document.querySelector('.page-right').classList.remove('hidden');
        }
        if (currentTPPage === 1)
            e.target.classList.add('hidden');
    });

    document.querySelector('.page-right').addEventListener('click', async(e) => {
        const newPosts = await HandleNextPosts(currentTPPage, document.querySelector('.posts'));
        tenPosts.push(...newPosts);
        currentTPPage++;
        document.querySelector('.page-left').classList.remove('hidden');
        if (newPosts.length < 11)
            e.target.classList.add('hidden');
    });
}

export const AddPost = (where, post, why = 'created') => {
    if (why === 'created' && tenPosts.length !== 10) {
        tenPosts.push(post);
        topPosts.push(post);
    } else {
        if (where === '10') tenPosts.push(post);
        else if (where === 'top') topPosts.push(post)
        else {
            tenPosts.push(post);
            topPosts.push(post);
        }
    }
}

export const ContactsPage = async() => {
    WriteContent(ContentHeader, 'Контакты');
    WriteContent(undefined, `<div class="contacts">
                                <div class="contacts-email">
                                    <i class="fa fa-envelope-o" aria-hidden="true"></i>
                                    <a href="mailto:anifor@inbox.ru" subject="От пользователя" target="_blank">Написать на почту</a>
                                </div>

                                <div class="contacts-github">
                                    <h2>Перейти на <i class="fa fa-github" aria-hidden="true"></i>-страницу разработчиков</h2>
                                    <a href="https://github.com/miron-developer" target="_blank" rel="noopener noreferrer">Miron-developer</a>
                                    <a href="https://github.com/mirask" target="_blank" rel="noopener noreferrer">MirasK</a>
                                </div>
                            </div>`);
    RemovePreloader();
}

// TODO: here filters
export const AdvancedSearchPage = async() => {
    WriteContent(ContentHeader, 'Расширенный поиск');
    WriteContent(undefined, `Расширенный поиск`);
    RemovePreloader();
}

export const SignInPage = async() => {
    RemovePreloader();
    if (GetUserStatus()) return Route('/');

    addCSS('sign');
    WriteContent(ContentHeader, 'Войти');
    WriteContent(undefined, `<div class="sign-in">
                                <form action="/sign/in" method="POST">
                                    <label>Логин или E-mail: <input type="text" name="login" placeholder="Логин или Email" required></label>
                                    <label>Пароль: <input type="password" name="password" placeholder="Пароль" required></label>
                                    <input type="submit" value="Отправить">
                                </form>
                            </div>
                            <div class="other-sign">
                                <div class="sign-up">Зарегистрироваться</div>
                                <div class="sign-restore">Восстановить пароль</div>
                            </div>`);

    document.querySelector('.sign-up').addEventListener('click', () => Route('/sign/up'));
    document.querySelector('.sign-restore').addEventListener('click', () => Route('/sign/restore'));
    document.querySelector('.sign-in form').addEventListener('submit', SignIn)
}

export const SignUpPage = async() => {
    RemovePreloader();
    if (GetUserStatus()) return Route('/');
    addCSS('sign');
    WriteContent(ContentHeader, 'Зарегистрироваться');
    WriteContent(undefined, `<div class="sign-up">
                                <form action="/sign/up" method="POST">
                                    <label>Никнейм: <input type="text" name="nickname" placeholder="Никнейм" required></label>
                                    <label>Возраст: <input type="number" name="age" placeholder="Возраст" required></label>
                                    <div>Пол:
                                        <label> Мужской <input type="radio" name="gender" value="Мужской"></label>
                                        <label> Женский <input type="radio" name="gender" value="Женский"></label>
                                    </div>
                                    <label>Имя: <input type="text" name="firstName" placeholder="Имя" required></label>
                                    <label>Фамилия: <input type="text" name="lastName" placeholder="Фамилия" required></label>
                                    <label>E-mail: <input type="email" name="email" placeholder="E-mail" required></label>
                                    <div class="sign-password-hint">Пароль должен содержать как минимум 8 символов: 1 и более с большой буквы, 1 и более с маленькой и 1 и более цифр</div>
                                    <label>Пароль: <input type="password" name="password" placeholder="Пароль" required></label>
                                    <label>Повторите пароль: <input type="password" name="repeatPassword" placeholder="Повторите пароль" required></label>
                                    <input type="submit" value="Отправить">
                                </form>
                            </div>`);
    document.querySelector('.sign-up form').addEventListener('submit', SingUp);
}

export const SaveUserPage = async() => {
    RemovePreloader();
    addCSS('sign');
    WriteContent(ContentHeader, 'Подтверждение регистрации');
    WriteContent(undefined, `<div class="save-user">
                                <h2 class="save-h2">Сообщение отправлено на почту</h2>
                                <form action="/sign/s/" method="POST">
                                    <label>Введите код: <input type="text" name="code" required></label>
                                    <input type="submit" value="Отправить">
                                </form>
                            </div>`);
    document.querySelector('.save-user form').addEventListener('submit', SaveUser);

    const code = window.location.pathname.split('/')[3];
    if (code !== '') {
        document.querySelector('.save-user input:first-child').value = code;
        SaveUser();
    }
}

export const RestorePasswordPage = async() => {
    RemovePreloader();
    addCSS('sign');
    WriteContent(ContentHeader, 'Подтверждение смены пароля');
    WriteContent(undefined, `<div class="restore-password">
                                <h2 class="restore-h2">Сообщение отправлено на почту</h2>
                                <form action="/sign/r/" method="POST">
                                    <label>Введите код: <input type="text" name="code" required></label>
                                    <div class="restore-password-hint">Пароль должен содержать как минимум 8 символов: 1 и более с большой буквы, 1 и более с маленькой и 1 и более цифр</div>
                                    <label>Введите новый пароль: <input type="password" name="password" required></label>
                                    <input type="submit" value="Отправить">
                                </form>
                            </div>`);
    document.querySelector('.restore-password form').addEventListener('submit', RestorePassword);

    const code = window.location.pathname.split('/')[3];
    if (code !== '') document.querySelector('.restore-password input:first-child').value = code;
}

export const RestorePage = async() => {
    RemovePreloader();
    if (GetUserStatus()) return Route('/');

    addCSS('sign');
    WriteContent(ContentHeader, 'Восстановление пароля');
    WriteContent(undefined, `<div class="restore">
                                <form action="/restore" method="POST">
                                    <label>E-mail: <input type="email" name="email" placeholder="E-mail" required></label>
                                    <input type="submit" value="Отправить">
                                    </form>
                            </div>`);
    document.querySelector('.restore form').addEventListener('submit', Restore)
}

export const PostPage = async() => {
    addCSS('post', 'posts');
    let post;
    const postID = parseInt(window.location.pathname.split('/')[2]);

    const isFromThisPosts = from => {
        const is = from.find(post => postID === post.id);
        if (is) return post = is;
        return false;
    }

    if (isFromThisPosts(myPosts)) null
    else if (isFromThisPosts(topPosts)) null
    else if (isFromThisPosts(tenPosts)) null
    else post = (await GetPost(postID))[0];

    const status = GetUserStatus();
    WriteContent(ContentHeader, post.title)
    WriteContent(undefined, `<section class="post-${post.id}">
                                <div class="post-header">
                                    <h2 class="post-title">${post.title}</h2>
                                    <div class="post-carma">
                                        <div class="post-carma-wrapper">
                                            <i class="fa fa-heart" aria-hidden="true"></i>
                                            <span class="post-carma-value">${post.carma}</span>
                                        </div>
                                        <div class="post-carma-actions">
                                            <div class="post-like">
                                                <i class="fa fa-thumbs-up" aria-hidden="true"></i>
                                            </div>
                                            <div class="post-dislike">
                                                <i class="fa fa-thumbs-down" aria-hidden="true"></i>
                                            </div>
                                        </div>
                                    </div>
                                </div>

                                <div class="post-tags-wrapper">
                                    <h2>Теги:</h2>
                                    <div class="post-tags">${GenTags(post.tags).slice(0, -1)}</div>
                                </div>

                                <div class="post-date-wrapper">
                                    <span class="post-date">Дата: ${post.date}</span>
                                    <span class="post-changed">${post.changed === 1 ? "(Изменено)" : ""}</span>
                                </div>

                                <div class="post-body">${post.body}</div>
                            </section>

                            ${status === 1?'<section class="your-comment"><h2>Ваш коментарий</h2></section>':''}
                            <section class="post-comments">
                                <h2>Коментарии поста</h2>
                                <div class="comments"></div>
                                <div class="also-comments">Показать еще коментарии</div>
                            </section>`);
    RemovePreloader();
    if (status === 1) MakeCommenting(post.id);
    StartComments(post.id);
    ShowComments(post.id);
    document.querySelector('.also-comments').addEventListener('click', () => ShowComments(post.id, '.post-comments .comments', 'post', 'beforeend'));
    document.querySelector('.post-like').addEventListener('click', HandleLike);
    document.querySelector('.post-dislike').addEventListener('click', HandleDislike);
}

export const ChangeProfilePage = async() => {
    RemovePreloader();
    if (!GetUserStatus()) return Route('/sign/in');

    addCSS('sign');
    WriteContent(ContentHeader, 'Изменить данные аккаунта');
    const user = GetUserData();
    WriteContent(undefined, `<div class="profile-change-avatar">
                                <h2>Нажмите на аватар, если хотите сменить</h2>
                                <div class="profile-avatar" title="file must be less than 20mb">
                                    <img src="/${user.photo}" alt="your-avatar">
                                </div>
                                <input name="oldphoto" value="${user.photo}" class="input-oldphoto" hidden>
                                <input type="file" name="avatar" class="avatar-change-input" accept=".jpg, .jpeg, .png, .gif" hidden>
                            </div>
                            <div class="change-profile-hint">Заполните то, что хотите изменить</div>
                            <div class="change-profile">
                                <form action="/change-profile" method="POST">
                                    <label>Новый никнейм: <input type="text" name="nick" placeholder="${user.nick}"></label>
                                    <label>Новый E-mail: <input type="email" name="email" placeholder="${user.email}"></label>
                                    <label>Новый возраст: <input type="number" name="age" placeholder="${user.age}"></label>
                                    <label>Новое имя: <input type="text" name="firstName" placeholder="${user.firstName}"></label>
                                    <label>Новая фамилия: <input type="text" name="lastName" placeholder="${user.lastName}"></label>
                                    <input type="submit" value="Отправить">
                                </form>
                            </div>`);

    document.querySelector('.profile-avatar').addEventListener('click', ChangeUserAvatar);
    document.querySelector('.change-profile form').addEventListener('submit', ChangeProfile)
}

export const CreatePostPage = async() => {
    RemovePreloader();
    if (!GetUserStatus()) return Route('/sign/in');

    addCSS('post');
    WriteContent(ContentHeader, 'Создать пост');
    WriteContent(undefined, `<div class="create-post"> 
                                <section class="tools">
                                    <h2>Инструменты</h2>

                                    <div class="tools-wrapper">
                                        <div class="tools-alignment">
                                            <div class="alignment-current current-left">
                                                <i class="fa fa-align-left" aria-hidden="true"></i> <span>Выравнивание</span>
                                            </div>
                                            <div class="alignment-aligns">
                                                <div class="align-left" title="по левому краю">
                                                    <i class="fa fa-align-left" aria-hidden="true"></i>
                                                    <span>по левому краю</span>
                                                </div>
                                                <div class="align-center" title="по центру">
                                                    <i class="fa fa-align-center" aria-hidden="true"></i>
                                                    <span>по центру</span>
                                                </div>
                                                <div class="align-right" title="по правому краю">
                                                    <span>по правому краю</span>
                                                    <i class="fa fa-align-right" aria-hidden="true"></i>
                                                </div>
                                                <div class="align-justify" title="по ширине">
                                                    <i class="fa fa-align-justify" aria-hidden="true"></i>
                                                    <span>по ширине</span>
                                                </div>
                                            </div>
                                        </div>

                                        <div class="tools-text-decoration">
                                            <div class="text-decoration-current current-normal">
                                                <i class="fa fa-font" aria-hidden="true"></i> <span>Текст</span>
                                            </div>
                                            <div class="decoration-decors">
                                                <div class="decors-bold" title="полу-жирное">
                                                    <i class="fa fa-bold" aria-hidden="true"></i>
                                                    <span>полу-жирное</span>
                                                </div>
                                                <div class="decors-italic" title="курсив">
                                                    <i class="fa fa-italic" aria-hidden="true"></i>
                                                    <span>курсив</span>    
                                                </div>
                                                <div class="decors-underline" title="подчеркнутое">
                                                    <i class="fa fa-underline" aria-hidden="true"></i>
                                                    <span>подчеркнутое</span>
                                                </div>
                                            </div>
                                        </div>

                                        <div class="font-sizing">
                                            <span title="Размер шрифта"><i class="fa fa-text-height" aria-hidden="true"></i></span>
                                            <select class="sizing-font-size">
                                                <option>8</option>
                                                <option>9</option>
                                                <option>10</option>
                                                <option>11</option>
                                                <option>12</option>
                                                <option selected>14</option>
                                                <option>16</option>
                                                <option>18</option>
                                                <option>20</option>
                                                <option>22</option>
                                                <option>24</option>
                                            </select>
                                        </div>
                                    
                                        <div class="tools-listing">
                                            <span><i class="fa fa-list" aria-hidden="true"></i> <span>Списки</span></span>
                                            <div class="listing-lists">
                                                <div class="list-ol" title="номерное">
                                                    <i class="fa fa-list-ol" aria-hidden="true"></i>
                                                    <span>номерное</span>
                                                </div>
                                                <div class="list-ul" title="символьное">
                                                    <i class="fa fa-list-ul" aria-hidden="true"></i>
                                                    <span>символьное</span>
                                                </div>
                                            </div>
                                        </div>
                                        <div class="elem-image" title="вставка изображения"><i class="fa fa-picture-o" aria-hidden="true"></i></div>
                                        <div class="elem-spoiler" title="вставка спойлера"><i class="fa fa-ban" aria-hidden="true"></i></div>
                                        <div class="elem-text" title="вставить текст">T</div>
                                        <div class="elem-citate" title="цитата или код"><i class="fa fa-code" aria-hidden="true"></i></div>
                                    </div>
                                </section>

                                <section class="post-body-wrapper">
                                    <label class="create-post-title">Тема(тайтл) поста:<input type="text" name="title" placeholder="Наруто Узумаки" required></label>
                                    <button class="generate-new-tag">Добавить тег</button>
                                    <div class="create-post-body"></div>
                                </section>

                                <section class="manage">
                                    <div class="save" title="сохранить и отправить"><i class="fa fa-floppy-o" aria-hidden="true"></i></div>
                                    <div class="delete-current-elem" title="удалить текущий элемент"><i class="fa fa-trash" aria-hidden="true"></i></div>
                                </section>
                            </div>`);

    let index = 1;
    GenerateDatalistTag();
    document.querySelector('.generate-new-tag').addEventListener('click', () => GenerateDatalistTag(++index));
    InitPE();
}

let myPosts = [];
let gettedMyPosts = false;
export const MyPostsPage = async() => {
    let currentMPPage = 1;
    RemovePreloader();
    if (!GetUserStatus()) return Route('/sign/in');
    addCSS('posts');

    WriteContent(ContentHeader, 'Мои посты');
    WriteContent(undefined, `<div class="my-posts">
                                <div class="posts"></div>
                            </div>
                            ${CreatePagination()}`);

    if (!gettedMyPosts) {
        myPosts = await GetPosts(0, 11, 'my')
        gettedMyPosts = true;
    }
    if (myPosts.length === 0) return WriteContent(document.querySelector('.content-body'), 'Постов нет');
    if (myPosts.length < 11) document.querySelector('.page-right').classList.add("hidden");
    WriteContent(document.querySelector('.posts'), GeneratePosts(myPosts.slice(0, 10)));
    HandlePosts();

    // handle pagination
    document.querySelector('.page-left').addEventListener('click', e => {
        currentMPPage--;
        if (currentMPPage > 0) {
            HandlePrevPosts(currentMPPage, document.querySelector('.posts'), myPosts.slice((currentMPPage - 1) * 10, currentMPPage * 10));
            document.querySelector('.page-right').classList.remove('hidden');
        }
        if (currentMPPage === 1)
            e.target.classList.add('hidden');
    });

    document.querySelector('.page-right').addEventListener('click', async(e) => {
        const newPosts = await HandleNextPosts(currentMPPage, document.querySelector('.posts'), "my");
        myPosts.push(...newPosts);
        currentMPPage++;
        document.querySelector('.page-left').classList.remove('hidden');
        if (newPosts.length < 11)
            e.target.classList.add('hidden');
    });
}

const checkIsUserOnline = nick => {
    const inter = setInterval(() => {
        let isUserExist;
        for (let user of OnlineUsers.onlines)
            if (user.Nickname === nick) isUserExist = user;

        if (isUserExist === undefined || !GetUserStatus()) {
            InformUser('user offline');
            clearInterval(inter);
            return Route('/');
        }
    }, 1000);
}

export const ChatPage = async() => {
    let currentLoadMessages = 1;
    RemovePreloader();
    if (!GetUserStatus()) return Route('/sign/in');

    const nick = window.location.pathname.split('/')[2];
    const users = Array.from(document.querySelectorAll('.user'));
    const user = users.find(user => user.childNodes[0].textContent === nick);
    if (!user) {
        InformUser("user not founded")
        return Route('/');
    }
    user.childNodes[1].textContent = '';
    user.childNodes[1].classList.remove('is-have-content');
    checkIsUserOnline(nick);

    let msgs = await GetMessages(nick);
    if (!msgs) return;
    msgs = msgs.reverse();

    const ID = GetUserData().id;
    const msgsDiv = msgs.reduce(
        (acc, msg) => {
            const date = msg.date.split(" ");
            return acc += `<div class="chat-msg ${msg.receiver===ID?'reverse':''}">
                                <span class="msg-date">${date[0]}\n${date[1]}</span>
                                <span class="msg-msg">${msg.body}</span>
                            </div>`;
        },
        ''
    );

    addCSS('chat', 'typing');
    removeCSS('sign');
    WriteContent(ContentHeader, 'Chat with ' + nick);
    WriteContent(undefined, `<div class="chat-msgs">${msgsDiv}</div>
                            <div class="chat-type">
                                <form action="/" class="send-ws">
                                    <input class="ws-body" type="text" required placeholder="text">
                                    <label class="send-file" title="send file">
                                        <i class="fa fa-file-image-o" aria-hidden="true"></i>
                                        <input class="send-file-input" type="file">
                                    </label>
                                    <label class="send-message" title="send message">
                                        <i class="fa fa-paper-plane" aria-hidden="true"></i>
                                        <input type="submit" class="send-message-input" accept=".jpg, .jpeg, .png, .gif" value="send">
                                    </label>
                                </form>
                            </div>`);

    // typing-in-progress
    document.querySelector('.ws-body').addEventListener('input', () => {
        StartTyping(nick);
        StopTyping(nick);
    });

    // file send
    document.querySelector('.send-file-input').addEventListener('change', e => {
        const file = e.target.files[0];
        SaveAsMsg(file, nick);
        SendFileOverWS(file, nick);
    });

    // ws send
    document.querySelector('.send-ws').addEventListener('submit', e => {
        e.preventDefault();
        const msg = document.querySelector('.ws-body').value;
        SendWSMessage(1, nick, msg);
        SendExStopTyping(nick);
        CreateMessage(msg);
        e.target.reset();
    });

    let currentPos;
    ResetLoad();
    document.querySelector('.chat-msgs').addEventListener('scroll', async(e) => {
        currentPos = e.target.scrollTop;
        if (currentPos === 0) {
            LazyLoad(nick, ID, e.target, currentLoadMessages * 10, 10);
            currentLoadMessages++;
        }
    });
}

export const UserProfilePage = async() => {
    RemovePreloader();
    if (!GetUserStatus()) return Route('/sign/in');
    const id = document.location.pathname.split('/')[2];
    const user = (await GetUser(id))[0];
    if (user === undefined) {
        InformUser("wrong user");
        return Route('/');
    }

    addCSS('sign', 'user');
    WriteContent(ContentHeader, `Профиль  ${user.nick}`);
    WriteContent(undefined, `  <div class="user-profile user-${user.id}">
                                    <div class="user-photo">
                                        <img src="/${user.photo}" alt="user-avatar">
                                    </div>

                                    <div class="user-nickname">Nickname: ${user.nick}</div>

                                    <div class="user-name">
                                        <div class="user-fname">First Name: ${user.firstName} &</div>
                                        <div class="user-lname">& Last Name: ${user.lastName}</div>
                                    </div>

                                    <div class="user-age">Age: ${user.age}</div>
                                    <div class="user.gender">Gender: ${user.gender}</div>

                                    <div class="user-carma">
                                        <div class="user-carma-wrapper">
                                            <i class="fa fa-heart" aria-hidden="true"></i>
                                            <span class="user-carma-value">${user.carma}</span>
                                        </div>
                                        <div class="user-carma-actions">
                                            <div class="user-like">
                                                <i class="fa fa-thumbs-up" aria-hidden="true"></i>
                                            </div>
                                            <div class="user-dislike">
                                                <i class="fa fa-thumbs-down" aria-hidden="true"></i>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                                <div class="start-chat">Chat with ${user.nick}</div>
                                `);


    document.querySelector('.user-like').addEventListener('click', HandleLike);
    document.querySelector('.user-dislike').addEventListener('click', HandleDislike);
    document.querySelector('.start-chat').addEventListener('click', () => Route('/chat/' + user.nick))
}