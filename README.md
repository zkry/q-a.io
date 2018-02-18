# q-a.io
For a 30 minute presentation with hundreds of listeners and only five minutes for question and answer, there's a better way than picking two lucky people. q-a.io is a (soon to be released) web app letting users submit questions and have them be voted on in real time. Let the questions with the most interest be asked and don't be interrupted with off topic questions.

# Building
The project has the API implemented in Golang and the frontend using the Vue framework. While it is still a work in progress, you can run the project by downloading the source code, going into the downloaded folder and running:

    cd public
    npm run build
    cd ..
    go run main.go

### Plan
#### CORE:
- [x] Add Close Room functionality
- [ ] Add self deleting rooms
- [x] Local Sorage ID saving
- [x] Depoy and connect frontend with backend
- [ ] Enable CI with ??? (Travis-CI)

#### WOULD BE GOOD:
- [x] Sort questions based on number of votes.
- [x] Observer sees votes when exits and comes back
- [x] Disable everything on observers side when the room is closed

#### Extra:
- [ ] Refactor Vue code
- [ ] Utilize websockets for reduced polling
- [ ] Get a code review
- [ ] Add email saving and end of session report
- [ ] Update UI:
    - [ ] Flame bar for hot votes
    - [ ] Colorful header
    - [ ] Usefull instructions and about
- [ ] GeoIP-security
- [ ] Refactor Go Code

#### Bugs:
- [ ] When server is reset, / route goes to manage page for room that doesn't exist
