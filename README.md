# SMB_Score_Bot
A Discord chat bot that gives time/score ILs for the games Super Monkey Ball, Super Monkey Ball 2, and Super Monkey Ball Deluxe.
Records are retreived from a designated Google Sheet (generally: https://docs.google.com/spreadsheets/d/1KoneeqJzheHFYapQ_JfyxL9sI0X8_BE7ZEVMZt0t0bI/edit)

# Chat Usage

Syntax Explanation:

    Anything inside () is required
    Anything inside [] is optional
    Values separated by | are alternative options. Only one of the options may be chosen (ex: !b10, not !ba10)

Stage IL

    use !(b|a|e|m)[x](<stageNumber>)
        b: beginner stages
        a: advanced stages
        e: expert stages
        m: master stages
        x: Used if this this an extra stage (ex: beginner extra)
        stageNumber: The number of the stage to retrieve
        
Story IL

    use !s(<world>)-(<floor>)
        s: Designated to look for a Story IL
        world: Which world to search in (ex: 1 = world 1)
        floor: Which level in the world to request (ex: 1 = floor 1)

# Command Line Usage
Various api keys and parameters are passed via command line. They can either be direct values or links to files.

    Use '-email="<EMAIL>"' To specify the Google Developer Credential email
    Use '-email-file="<EMAILFILE>" To specify a file the Google Developer Credential email is in
    Use '-privatekey="<>"' to specify a file the Google API private key is in
    Use '-disctoken="<TOKEN>"' to specify a file that the Discord API token is in
    Use '-sheet="<SHEETID>"' to the sheet id containing the IL information
    Use '-sheet-file="<SHEETFILE>"' to specify a file that sheet id conatining IL information is in
    Use '-test=""' to specify test mode (only responds in the test discord server)