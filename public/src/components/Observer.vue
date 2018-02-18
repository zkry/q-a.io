<template>
  <div class="container">
    <h1>Unnamed Project</h1>
    <transition name="fadefast">
      <div class="error-msg" v-show="errorActive">
        <a class="cancel" @click="hideError">&times;</a> {{ errorMsg }}
      </div>
    </transition>
    <h2>{{ roomName }}<span class="info-text" v-if="closedRoom"> closed</span></h2>
    <template v-if="validRoom">
      <h4>Questions:</h4>
      <div class="question-box-container">
        Ask a question:
         <input class="question-box" type="text" v-model="askingQuestion" :disabled="closedRoom">
         <span @click="sendQuestion" class="envelope" > &#9993;</span>
      </div>
      <ul class="question-container">
        <li class="question-item" v-for="key in getQuestionKeys" :key="key" :class="{ 'closed-question' : closedRoom }">
          <span class="vote-arrow" @click="vote(questions[key].id, '1')" v-bind:class="[questions[key].myVote === '1' ? 'active-vote' : '']">▲</span>
           <span class="vote-ct" >{{ questions[key].vote }}</span>
           <span class="vote-arrow" @click="vote(questions[key].id, '-1')" v-bind:class="[questions[key].myVote === '-1' ? 'active-vote' : '']">▼</span>
           {{ questions[key].q }}
        </li>
      </ul>
    </template>
  </div>
</template>

<script>
import axios from 'axios'

export default {
  name: 'Observer',
  data: function () {
    return {
      errorMsg: '',
      errorActive: false,
      roomName: '',
      validRoom: true,
      closedRoom: false,
      questions: {},
      askingQuestion: '',
      sortMode: 'NORMAL'
    }
  },
  computed: {
    getQuestionKeys () {
      let baseKeys = Object.keys(this.questions)
      if (this.sortMode === 'NORMAL') {
        return baseKeys
      } else if (this.sortMode === 'SORTED') {
        baseKeys.sort((a, b) => {
          return this.questions[a].vote < this.questions[b].vote
        })
        return baseKeys
      } else if (this.sortMode === 'SORTED_REVERSE') {
        baseKeys.sort((a, b) => {
          return this.questions[a].vote > this.questions[b].vote
        })
        return baseKeys
      }
      return baseKeys
    }
  },
  created: function () {
    // Register then get questions
    this.roomName = this.$route.params.roomName

    // Check if we have an ID saved. We will send our 'old' id to the server,
    // if it is still valid we will get the same one back, else get a new one.
    let oldID = localStorage.getItem(this.roomName + ':userID')
    const formData = {
      uID: oldID
    }

    axios({
      method: 'post',
      url: '/api/v1/register/' + this.roomName,
      data: formData,
      params: formData,
      headers: {
        'Content-type': 'application/x-www-form-urlencoded; charset=UTF-8'
      }
    })
      .then(res => {
        this.userID = res.data.id
        // If our old ID is the same we can reuse other data
        if (oldID === res.data.id) {
          this.loadVotes()
        }
        console.log('Registered with id ', this.userID)
        localStorage.setItem(this.roomName + ':userID', this.userID)
        this.getQuestions()
        setInterval(() => {
          this.getQuestions()
        }, 2000)
      }).catch(err => {
        console.log(err)
        if (err.response && err.response.data.status === 'Room does not exist') {
          this.roomName = `Room ${this.roomName} does not exist`
          this.validRoom = false
        }
        // TODO: Display no-such-room-exists message
      })
  },
  methods: {
    sendQuestion () {
      console.log('Sending Message')
      if (this.askingQuestion.length <= 5) {
        this.showError('Message is too short to be valid.')
        return
      }
      const formData = {
        uID: this.userID,
        question: this.askingQuestion
      }
      axios({
        method: 'post',
        url: '/api/v1/publishQuestion/' + this.roomName,
        data: formData,
        params: formData,
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded; charset=UTF-8'
        }
      })
        .then(res => {
          console.log('Question added')
          this.askingQuestion = ''
          this.getQuestions()
        })
        .catch(err => {
          this.showError(err.response.data.status)
        })
    },
    vote (id, btn) {
      let voteVal = btn
      if (this.questions[id].myVote === btn) {
        voteVal = '0'
      }
      const formData = {
        qID: id,
        val: voteVal,
        uID: this.userID
      }
      axios({
        method: 'post',
        url: '/api/v1/vote/' + this.roomName,
        data: formData,
        params: formData,
        headers: {
          'Content-type': 'application/x-www-form-urlencoded; charset=UTF-8'
        }
      })
        .then(res => {
          this.questions[id].myVote = voteVal
          // We should also save our vote
          this.saveVotes()
          this.$forceUpdate()
          this.getQuestions()
        })
        .catch(err => {
          console.log(err.response.data)
        })
    },
    saveVotes () {
      console.log('DEBUG: saving votes')
      let myVotes = {}
      // TODO: Create specialized objet for my votes to avoid this iteration
      for (var key in this.questions) {
        if (this.questions[key].myVote === '-1' || this.questions[key].myVote === '1') {
          myVotes[key] = this.questions[key].myVote
        }
      }
      localStorage.setItem(this.roomName + ':votes', JSON.stringify(myVotes))
    },
    loadVotes () {
      this.getQuestions().then(() => {
        console.log('DEBUG: loading votes')
        if (localStorage.getItem(this.roomName + ':votes') !== null) {
          let myVotes = JSON.parse(localStorage.getItem(this.roomName + ':votes'))
          for (var key in myVotes) {
            this.questions[key].myVote = myVotes[key]
          }
        }
      })
    },
    hideError () {
      this.errorActive = false
    },
    showError (errorMsg) {
      this.errorMsg = errorMsg
      this.errorActive = true
    },
    getQuestions () {
      return axios({
        method: 'get',
        url: '/api/v1/getQuestions/' + this.roomName,
        headers: {
          'Content-type': 'application/x-www-form-urlencoded; charset=UTF-8'
        }
      })
        .then(res => {
          // TODO: Add res → this.questions
          for (let i = 0; i < res.data.questions.length; i++) {
            let id = res.data.questions[i].id
            if (!this.questions.hasOwnProperty(id)) {
              this.$set(this.questions, id, {q: res.data.questions[i].q, vote: res.data.questions[i].vote, id: id, myVote: '0'})
            } else if (this.questions[id].vote !== res.data.questions[i].vote) {
              this.$set(this.questions, id, {q: res.data.questions[i].q, vote: res.data.questions[i].vote, id: id, myVote: this.questions[id].myVote})
            }
          }
          this.closedRoom = res.data.is_closed
        })
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
h1, h2 {
  font-weight: normal;
}
ul {
  list-style-type: none;
  padding: 0;
}
li {
  margin: 0 10px;
}
a {
  color: #42b983;
}
p {
  margin: 0;
  padding: 0;
}

.cancel {
  color: black;
  font-weight: bold;
  cursor: pointer;
}

.error-msg {
  position: absolute;
  background-color: #e74c3c;
  top: 15px;
  right: 15px;
  padding: 5px;
  border-radius: 4px;
  cursor: grab;
  width: 200px;
}

.go-btn {
  margin: 0px;
  cursor: pointer;
}
.container {
}
.input {
  margin: 0;
  padding: 0;
}

.envelope {
  position: relative;
  font-size: xx-large;
  top: 5px;
  cursor: pointer;
}

.info-text {
  font-size: x-small;
  color: grey;
}

.question-box-container{
  width: 90%;
  margin: auto;
}

.question-box{
  width: 60%;
}

.question-container {
  margin: 5px 50px 0px 50px;
  text-align: left;
  list-style-type: none;
}

.question-item {
  white-space: nowrap;
}

.closed-question {
  /* background-color: #95a5a6; */
}

.vote-arrow {
  cursor: pointer;
}

.vote-arrow::-moz-selection { background:transparent; }
.vote-arrow::selection { background:transparent; }
.vote-ct::-moz-selection { background:transparent; }
.vote-ct::selection { background:transparent; }

.active-vote {
  color: #d35400;
}

.vote-ct {
  /* make width fixed ??? */
}
/* fadefast */
.fadefast-enter {
  opacity: 0;
}
.fadefast-enter-active {
  transition: opacity 0.5s;
}
.fadefast-leave {

}
.fadefast-leave-active {
  transition: opacity 0.5s;
  opacity: 0;
}

/* fade */
.fade-enter {
  opacity: 0;
}
.fade-enter-active {
  transition: opacity 1s;
  animation: slide-in 1s ease-out forwards;
}
.fade-leave {

}
.fade-leave-active {
  transition: opacity 1s;
  opacity: 0;
  animation: slide-out 1s ease-out forwards;
}

@keyframes slide-in {
  from {
    transform: translateY(20px);
  }
  to {
    transform: translateY(0px);
  }
}
@keyframes slide-out {
  from {
    transform: translateY(0px);
  }
  to {
    transform: translateY(20px);
  }
}
</style>
