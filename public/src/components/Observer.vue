<template>
  <div class="container">
    <h1>Unnamed Project</h1>
    <transition name="fadefast">
      <div class="error-msg" v-show="errorActive">
        <a class="cancel" @click="hideError">&times;</a> {{ errorMsg }}
      </div>
    </transition>
    <h2>{{ roomName }}</h2>
    <template v-if="validRoom">
      <h4>Questions:</h4>
      <ul class="question-container">
        <li class="question-item" v-for="question in questions" :key="question.id">
          <span class="vote-arrow" @click="vote(question.id, '1')" v-bind:class="[question.myVote === '1' ? 'active-vote' : '']">▲</span>
           <span class="vote-ct" >{{ question.vote }}</span>
           <span class="vote-arrow" @click="vote(question.id, '-1')" v-bind:class="[question.myVote === '-1' ? 'active-vote' : '']">▼</span>
           {{ question.q }}
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
      questions: {}
    }
  },
  computed: {
  },
  created: function () {
    // Register then get questions
    this.roomName = this.$route.params.roomName
    axios({
      method: 'post',
      url: 'http://localhost:8080/register/' + this.roomName,
      headers: {
        'Content-type': 'application/x-www-form-urlencoded; charset=UTF-8'
      }
    })
      .then(res => {
        this.userID = res.data.id
        console.log('Registered with id ', this.userID)
        this.getQuestions()
        setInterval(() => {
          this.getQuestions()
        }, 2000)
      }).catch(err => {
        console.log(err.response.data)
        if (err.response.data.status === 'Room does not exist') {
          this.roomName = `Room ${this.roomName} does not exist`
          this.validRoom = false
        }
        // TODO: Display no-such-room-exists message
      })
  },
  methods: {
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
        url: 'http://localhost:8080/vote/' + this.roomName,
        data: formData,
        params: formData,
        headers: {
          'Content-type': 'application/x-www-form-urlencoded; charset=UTF-8'
        }
      })
        .then(res => {
          this.questions[id].myVote = voteVal
          this.$forceUpdate()
          this.getQuestions()
        })
        .catch(err => {
          console.log(err.response.data)
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
      axios({
        method: 'get',
        url: 'http://localhost:8080/getQuestions/' + this.roomName,
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

.question-container {
  margin: 50px;
  text-align: left;
  list-style-type: none;
}

.question {
  display: block;
  white-space: nowrap;
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
