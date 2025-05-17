Getting Started
Datastar brings the functionality provided by libraries like Alpine.js (frontend reactivity) and htmx (backend reactivity) together, into one cohesive solution. It’s a lightweight, extensible framework that allows you to:

Manage state and build reactivity into your frontend using HTML attributes.
Modify the DOM and state by sending events from your backend.
With Datastar, you can build any UI that a full-stack framework like React, Vue.js or Svelte can, but with a much simpler, hypermedia-driven approach.

We're so confident that Datastar can be used as a JavaScript framework replacement that we challenge anyone to find a use-case for a web app that Datastar cannot be used to build!
Installation#
The quickest way to use Datastar is to include it in your HTML using a script tag hosted on a CDN.

<script type="module" src="https://cdn.jsdelivr.net/gh/starfederation/datastar@v1.0.0-beta.11/bundles/datastar.js"></script>
If you prefer to host the file yourself, download your own bundle using the bundler, then include it from the appropriate path.

<script type="module" src="/path/to/datastar.js"></script>
You can alternatively install Datastar via npm. We don’t recommend this for most use-cases, as it requires a build step, but it can be useful for legacy frontend projects.

npm install @starfederation/datastar
Data Attributes#
At the core of Datastar are data-* attributes (hence the name). They allow you to add reactivity to your frontend in a declarative way, and to interact with your backend.

Datastar uses signals to manage state. You can think of signals as reactive variables that automatically track and propagate changes in expressions. They can be created and modified using data attributes on the frontend, or events sent from the backend. Don’t worry if this sounds complicated; it will become clearer as we look at some examples.

The Datastar VSCode extension and IntelliJ plugin provided autocompletion for all data-* attributes.
data-bind#
Datastar provides us with a way to set up two-way data binding on an element using the data-bind attribute, which can be placed on any HTML element on which data can be input or choices selected from (input, textarea, select, checkbox and radio elements, as well as web components).

<input data-bind-input />
This creates a new signal that can be called using $input, and binds it to the element’s value. If either is changed, the other automatically updates.

An alternative syntax, in which the value is used as the signal name, is also available. This can be useful depending on the templating language you are using.

<input data-bind="input" />
data-text#
To see this in action, we can use the data-text attribute.

<input data-bind-input />
<div data-text="$input">
  I will be replaced with the contents of the input signal
</div>
Input:
Output:
This sets the text content of an element to the value of the signal $input. The $ prefix is required to denote a signal.

Note that data-* attributes are evaluated in the order they appear in the DOM, so the data-text attribute must come after the data-bind attribute. See the attribute plugins reference for more information.

The value of the data-text attribute is a Datastar expression that is evaluated, meaning that we can use JavaScript in it.

<input data-bind-input />
<div data-text="$input.toUpperCase()">
  Will be replaced with the uppercase contents of the input signal
</div>
Input:
Output:
data-computed#
The data-computed attribute creates a new signal that is computed based on a reactive expression. The computed signal is read-only, and its value is automatically updated when any signals in the expression are updated.

<input data-bind-input />
<div data-computed-repeated="$input.repeat(2)">
    <div data-text="$repeated">
        Will be replaced with the contents of the repeated signal
    </div>
</div>
This results in the $repeated signal’s value always being equal to the value of the $input signal repeated twice. Computed signals are useful for memoizing expressions containing other signals.

Input:
Output:
data-show#
The data-show attribute can be used to show or hide an element based on whether an expression evaluates to true or false.

<input data-bind-input />
<button data-show="$input != ''">Save</button>
This results in the button being visible only when the input is not an empty string (this could also be written as !input).

Input:
Output:
data-class#
The data-class attribute allows us to add or remove a class to or from an element based on an expression.

<input data-bind-input />
<button data-class-hidden="$input == ''">Save</button>
If the expression evaluates to true, the hidden class is added to the element; otherwise, it is removed.

Input:
Output:
The data-class attribute can also be used to add or remove multiple classes from an element using a set of key-value pairs, where the keys represent class names and the values represent expressions.

<button data-class="{hidden: $input == '', 'font-bold': $input == 1}">Save</button>
data-attr#
The data-attr attribute can be used to bind the value of any HTML attribute to an expression.

<input data-bind-input />
<button data-attr-disabled="$input == ''">Save</button>
This results in a disabled attribute being given the value true whenever the input is an empty string.

Input:
Output:
The data-attr attribute can also be used to set the values of multiple attributes on an element using a set of key-value pairs, where the keys represent attribute names and the values represent expressions.

<button data-attr="{disabled: $input == '', title: $input}">Save</button>
data-signals#
So far, we’ve created signals on the fly using data-bind and data-computed. All signals are merged into a global set of signals that are accessible from anywhere in the DOM.

We can also create signals using the data-signals attribute.

<div data-signals-input="1"></div>
Using data-signals merges one or more signals into the existing signals. Values defined later in the DOM tree override those defined earlier.

Signals can be namespaced using dot-notation.

<div data-signals-form.input="2"></div>
The data-signals attribute can also be used to merge multiple signals using a set of key-value pairs, where the keys represent signal names and the values represent expressions.

<div data-signals="{input: 1, form: {input: 2}}"></div>
data-on#
The data-on attribute can be used to attach an event listener to an element and execute an expression whenever the event is triggered.

<input data-bind-input />
<button data-on-click="$input = ''">Reset</button>
This results in the $input signal’s value being set to an empty string whenever the button element is clicked. This can be used with any valid event name such as data-on-keydown, data-on-mouseover, etc.

Input:
Output:
So what else can we do now that we have declarative signals and expressions? Anything we want, really!

See if you can follow the code below based on what you’ve learned so far, before trying the demo.

<div
  data-signals="{response: '', answer: 'bread'}"
  data-computed-correct="$response.toLowerCase() == $answer"
>
  <div id="question">What do you put in a toaster?</div>
  <button data-on-click="$response = prompt('Answer:') ?? ''">BUZZ</button>
  <div data-show="$response != ''">
    You answered “<span data-text="$response"></span>”.
    <span data-show="$correct">That is correct ✅</span>
    <span data-show="!$correct">
      The correct answer is “
      <span data-text="$answer"></span>
      ” 🤷
    </span>
  </div>
</div>
What do you put in a toaster?
We’ve just scratched the surface of frontend reactivity. Now let’s take a look at how we can bring the backend into play.

Backend Setup#
Datastar uses Server-Sent Events (SSE) to stream zero or more events from the web server to the browser. There’s no special backend plumbing required to use SSE, just some syntax. Fortunately, SSE is straightforward and provides us with some advantages.

First, set up your backend in the language of your choice. Familiarize yourself with sending SSE events, or use one of the backend SDKs to get up and running even faster. We’re going to use the SDKs in the examples below, which set the appropriate headers and format the events for us.

The following code would exist in a controller action endpoint in your backend.

import (datastar "github.com/starfederation/datastar/sdk/go")

// Creates a new `ServerSentEventGenerator` instance.
sse := datastar.NewSSE(w,r)

// Merges HTML fragments into the DOM.
sse.MergeFragments(
    `<div id="question">What do you put in a toaster?</div>`
)

// Merges signals into the signals.
sse.MergeSignals([]byte(`{response: '', answer: 'bread'}`))
The mergeFragments() method merges the provided HTML fragment into the DOM, replacing the element with id="question". An element with the ID question must already exist in the DOM.

The mergeSignals() method merges the response and answer signals into the frontend signals.

With our backend in place, we can now use the data-on-click attribute to trigger the @get() action, which sends a GET request to the /actions/quiz endpoint on the server when a button is clicked.

<div
  data-signals="{response: '', answer: ''}"
  data-computed-correct="$response.toLowerCase() == $answer"
>
  <div id="question"></div>
  <button data-on-click="@get('/actions/quiz')">Fetch a question</button>
  <button
    data-show="$answer != ''"
    data-on-click="$response = prompt('Answer:') ?? ''"
  >
    BUZZ
  </button>
  <div data-show="$response != ''">
    You answered “<span data-text="$response"></span>”.
    <span data-show="$correct">That is correct ✅</span>
    <span data-show="!$correct">
      The correct answer is “<span data-text="$answer"></span>” 🤷
    </span>
  </div>
</div>
Now when the Fetch a question button is clicked, the server will respond with an event to modify the question element in the DOM and an event to modify the response and answer signals. We’re driving state from the backend!

data-indicator#
The data-indicator attribute sets the value of a signal to true while the request is in flight, otherwise false. We can use this signal to show a loading indicator, which may be desirable for slower responses.

<div id="question"></div>
<button
  data-on-click="@get('/actions/quiz')"
  data-indicator-fetching
>
  Fetch a question
</button>
<div data-class-loading="$fetching" class="indicator"></div>
The data-indicator attribute can also be written with signal name in the attribute value.

<button
  data-on-click="@get('/actions/quiz')"
  data-indicator="fetching"
>
We’re not limited to just GET requests. Datastar provides backend plugin actions for each of the methods available: @get(), @post(), @put(), @patch() and @delete().

Here’s how we could send an answer to the server for processing, using a POST request.

<button data-on-click="@post('/actions/quiz')">
  Submit answer
</button>
One of the benefits of using SSE is that we can send multiple events (HTML fragments, signal updates, etc.) in a single response.

sse.MergeFragments(`<div id="question">...</div>`)
sse.MergeFragments(`<div id="instructions">...</div>`)
sse.MergeSignals([]byte(`{answer: '...'}`))
sse.MergeSignals([]byte(`{prize: '...'}`))
Actions#
Actions in Datastar are helper functions that are available in data-* attributes and have the syntax @actionName(). We already saw the backend plugin actions above. Here are a few other useful actions.

@setAll()#
The @setAll() action sets the value of all matching signals to the expression provided in the second argument. The first argument can be one or more space-separated paths in which * can be used as a wildcard.

<button data-on-click="@setAll('foo.*', $bar)"></button>
This sets the values of all signals namespaced under the foo signal to the value of $bar. This can be useful for checking multiple checkbox fields in a form, for example:

<input type="checkbox" data-bind-checkboxes.checkbox1 /> Checkbox 1
<input type="checkbox" data-bind-checkboxes.checkbox2 /> Checkbox 2
<input type="checkbox" data-bind-checkboxes.checkbox3 /> Checkbox 3
<button data-on-click="@setAll('checkboxes.*', true)">Check All</button>
@toggleAll()#
The @toggleAll() action toggles the value of all matching signals. The first argument can be one or more space-separated paths in which * can be used as a wildcard.

<button data-on-click="@toggleAll('foo.*')"></button>
This toggles the values of all signals namespaced under the foo signal (to either true or false). This can be useful for toggling multiple checkbox fields in a form, for example:

<input type="checkbox" data-bind-checkboxes.checkbox1 /> Checkbox 1
<input type="checkbox" data-bind-checkboxes.checkbox2 /> Checkbox 2
<input type="checkbox" data-bind-checkboxes.checkbox3 /> Checkbox 3
<button data-on-click="@toggleAll('checkboxes.*')">Toggle All</button>
View the reference overview.