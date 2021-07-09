-- my public post
INSERT INTO Posts VALUES(
    null,
    "post",
    "test post",
    "post body",
    1588670115000,
    "public",
    "",
    1,
    1,
    null
);

-- my private post
INSERT INTO Posts VALUES(
    null,
    "post",
    "test post",
    "post body",
    1588670115000,
    "private",
    "",
    0,
    1,
    null
);

-- my almost_private post
INSERT INTO Posts VALUES(
    null,
    "post",
    "test post",
    "post body",
    1588670115000,
    "almost_private",
    "2,3",
    0,
    1,
    null
);

-- my following user public post
INSERT INTO Posts VALUES(
    null,
    "post",
    "test post",
    "post body",
    1588670115000,
    "public",
    "",
    0,
    2,
    null
);

-- my following user private post
INSERT INTO Posts VALUES(
    null,
    "post",
    "test post",
    "post body",
    1588670115000,
    "private",
    "",
    0,
    2,
    null
);

-- my following user almost_private post
INSERT INTO Posts VALUES(
    null,
    "post",
    "test post",
    "post body",
    1588670115000,
    "almost_private",
    "1,3",
    0,
    2,
    null
);

-- no following user public post
INSERT INTO Posts VALUES(
    null,
    "post",
    "test post",
    "post body",
    1588670115000,
    "public",
    "",
    0,
    5,
    null
);

-- no following user private post
INSERT INTO Posts VALUES(
    null,
    "post",
    "test post",
    "post body",
    1588670115000,
    "private",
    "",
    0,
    5,
    null
);

-- no following user almost_private post
INSERT INTO Posts VALUES(
    null,
    "post",
    "test post",
    "post body",
    1588670115000,
    "almost_private",
    "2,3",
    0,
    5,
    null
);

-- my following group public post
INSERT INTO Posts VALUES(
    null,
    "post",
    "test post",
    "post body",
    1588670115000,
    "public",
    "",
    0,
    null,
    1
);

-- my following group private post
INSERT INTO Posts VALUES(
    null,
    "post",
    "test post",
    "post body",
    1588670115000,
    "private",
    "",
    0,
    null,
    1
);

-- my following group almost_private post
INSERT INTO Posts VALUES(
    null,
    "post",
    "test post",
    "post body",
    1588670115000,
    "almost_private",
    "2",
    0,
    null,
    1
);

-- no following group public post
INSERT INTO Posts VALUES(
    null,
    "post",
    "test post",
    "post body",
    1588670115000,
    "public",
    "",
    0,
    null,
    4
);

-- no following group private post
INSERT INTO Posts VALUES(
    null,
    "post",
    "test post",
    "post body",
    1588670115000,
    "private",
    "",
    0,
    null,
    4
);

-- no following user almost_private post
INSERT INTO Posts VALUES(
    null,
    "post",
    "test post",
    "post body",
    1588670115000,
    "almost_private",
    "2,3",
    0,
    null,
    4
);