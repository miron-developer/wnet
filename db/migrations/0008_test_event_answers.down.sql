DELETE FROM EventAnswers WHERE 
    eventID = (SELECT id FROM Events WHERE title = "test event" LIMIT 1);