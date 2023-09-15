export default function Home() {
  return (
    <body>
      <script src="https://unpkg.com/htmx.org"></script>
      <script src="https://unpkg.com/htmx.org/dist/ext/client-side-templates.js"></script>
      <script src="https://unpkg.com/mustache@latest"></script>

      <div hx-ext="client-side-templates">
        <button hx-post="/api/insert" mustache-template="insert-template" hx-target="#insert-record" hx-swap="innerHTML">
          Insert
        </button>
        <button hx-get="/api/fetch" mustache-template="fetch-template" hx-target="#fetch-records" hx-swap="innerHTML">
          Fetch
        </button>

        <template id="insert-template" dangerouslySetInnerHTML={{ __html: `
          Time to insert a record: {{latency}}
        `}}/>

        <template id="fetch-template" dangerouslySetInnerHTML={{ __html: `
          Time to fetch records: {{latency}}
          <br>
          {{#records}}
          {{id}} - {{value}}<br>
          {{/records}}
        `}}/>

        <div id="insert-record">
        </div>

        <div id="fetch-records">
        </div>
      </div>
    </body>
  )
}
