import SwaggerUI from "swagger-ui-react"
import "swagger-ui-react/swagger-ui.css"
  
export default function API() {
return (
    <main style={{ padding: "1rem 0" }}>
      <SwaggerUI url="/api/swagger.json" />
    </main>
    )
}


