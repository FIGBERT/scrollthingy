import gleam/int
import lustre
import lustre/effect.{type Effect}
import lustre/element.{type Element}
import lustre/element/html
import plinth/browser/event.{type Event}
import plinth/browser/window

type Model =
  Int

type Msg {
  Wheel(delta: Int)
}

@external(javascript, "./scroll.mjs", "delta_from_event")
fn delta_from_event(evt: Event(a)) -> Int

fn init(_args: Nil) -> #(Model, Effect(Msg)) {
  #(
    0,
    effect.from(fn(dispatch) {
      window.add_event_listener("wheel", fn(evt) {
        dispatch(Wheel(delta_from_event(evt)))
      })
    }),
  )
}

fn update(model: Model, msg: Msg) -> #(Model, Effect(Msg)) {
  case msg {
    Wheel(delta) -> {
      let model = case delta {
        less if less < 0 -> model - 1
        more if more > 0 -> model + 1
        _ -> model
      }
      #(model, effect.none())
    }
  }
}

fn view(model: Model) -> Element(Msg) {
  html.p([], [html.text(int.to_string(model))])
}

pub fn main() -> Nil {
  let assert Ok(_) =
    lustre.application(init, update, view)
    |> lustre.start("body", Nil)
  Nil
}
