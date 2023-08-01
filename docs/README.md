# Pessimism Documentation

This directory contains the english specs for the Pessimism application. 

## Contents
- [Architecture](architecture/architecture.markdown)
- [JSON-RPC API](swaggerdoc.html)
- [ETL Subsystem](architecture/etl.markdown)
- [Engine Subsystem](architecture/engine.markdown)
- [Alerting Subsystem](architecture/alerting.markdown)
- [Heuristics](heuristics.markdown)
- [Telemetry](telemetry.markdown)

## GitHub Pages
The Pessimism documentation is hosted on GitHub Pages. To view the documentation, please visit [https://base-org.github.io/pessimism](https://base-org.github.io/pessimism/architecture). 
 

## Contributing
If you would like to contribute to the Pessimism documentation, please advise the guidelines stipulated in the [CONTRIBUTING.md](../CONTRIBUTING.md) file __before__ submitting a pull request.


## Running Docs Website Locally

### Prerequisites
- Ensure that you have installed the latest version of ruby on your machine following steps located [here](https://www.ruby-lang.org/en/documentation/installation/).
- Installing ruby should also install the ruby bundler which is used to install dependencies located in the [Gemfile](Gemfile)

### Local Testing
To run the documentation website locally, ensure you have followed the prerequisite steps, then do the following
1. Install dependencies via `bundle install`
2. Run `bundle exec jekyll serve`
3. You should now see a localhost version of documentation for the website!

## Creating Diagrams in GitHub Pages

It is important to note that you cannot simply write a mermaid diagram as you normally would with markdown and expect the diagram to be properly rendered via GitHub pages. To enable proper GitHub pages rendering, follow the recommended steps below:
1. Implement your diagram in markdown using the ` ```mermaid\n` key
2. Once done with implementing the diagram, ff you have not already, import the mermaid.js library via the following 
    ```
   {% raw %}
    <script src="https://cdn.jsdelivr.net/npm/mermaid@10.3.0/dist/mermaid.min.js"></script>
    {% endraw %}
   ```
3. Delete the ` ```mermaid ` key and replace it with 
   ```
   {% raw %}
    <div class="mermaid">
        --- diagram implementation here ---
    </div> 
   {% endraw %}
   
4. Done! To make sure this renders correctly, you can run `bundle exec jekyll serve` to view your changes.