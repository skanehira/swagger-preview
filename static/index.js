const app = new Vue({
  el: '#app',
  data: {
    ui: {},
    ws: {},
    fileType: "",
  },
  methods: {
  },
  mounted(){
    this.ws = new WebSocket("ws://localhost:8080/ws")

    this.ws.onopen = () => {
      this.ws.send("")
    }

    let isFirst = false
    this.ws.onmessage = (ev) => {
      const resp = JSON.parse(ev.data)
      if (!isFirst) {
        this.fileType = resp.fileType
        if (resp.fileType === "swagger") {
          this.ui = SwaggerUIBundle({
            url : resp.fileName,
            dom_id: '#ui',
            deepLinking: true,
            presets: [
              SwaggerUIBundle.presets.apis,
              SwaggerUIStandalonePreset
            ],
            plugins: [
              SwaggerUIBundle.plugins.DownloadUrl
            ],
            layout: "StandaloneLayout"
          })
        }

        isFirst = true
        return
      }

      console.log("updating...");
      if (this.fileType === "swagger") {
        this.ui.specActions.updateSpec(resp.message)
      } else if (this.fileType === "markdown") {

      } else if (this.fileType === "asciidoc") {

      }
    }

    this.ws.onerr = (err) => {
      console.log(err)
    }

    window.onbeforeunload = (ev) => {
      this.ws.close()
      console.log(ev)
    }
  }
})

