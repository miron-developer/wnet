-- my event
INSERT INTO Events VALUES(
    null,
    "event",
    "test event",
    "test event about",
    1588670115000,
    1,
    null
);

-- my group event
INSERT INTO Events VALUES(
    null,
    "event",
    "test event",
    "test event about",
    1588670115000,
    null,
    1
);

-- following user event
INSERT INTO Events VALUES(
    null,
    "event",
    "test event",
    "test event about",
    1588670115000,
    2,
    null
);

-- following group event
INSERT INTO Events VALUES(
    null,
    "event",
    "test event",
    "test event about",
    1588670115000,
    null,
    2
);

-- no following user event
INSERT INTO Events VALUES(
    null,
    "event",
    "test event",
    "test event about",
    1588670115000,
    5,
    null
);

-- no following group event
INSERT INTO Events VALUES(
    null,
    "event",
    "test event",
    "test event about",
    1588670115000,
    null,
    3
);