use cosmic::app::Core;
use cosmic::iced::wayland::popup::{destroy_popup, get_popup};
use cosmic::iced::window::Id;
use cosmic::iced::{Command, Limits};
use cosmic::iced_futures::Subscription;
use cosmic::iced_runtime::core::window;
use cosmic::iced_style::application;
use cosmic::widget;
use cosmic::{Element, Theme};
{{if .Config -}} 
use cosmic::cosmic_config;
use crate::config::{Config, CONFIG_VERSION};
{{- end}}
pub const ID: &str = "{{.ID}}";

#[derive(Default)]
pub struct Window {
    core: Core,
    popup: Option<Id>,
    {{if .Example -}} example_row: bool, {{- end}}
    {{if .Config -}}
    config: Config,
    #[allow(dead_code)]
    config_handler: Option<cosmic_config::Config>,
    {{- end}}
}

#[derive(Clone, Debug)]
pub enum Message {
    {{if .Config -}} Config(Config), {{- end}}
    TogglePopup,
    PopupClosed(Id),
    {{if .Example -}} ToggleExampleRow(bool),{{- end}}
}
{{if .Config -}}
#[derive(Clone, Debug)]
pub struct Flags {
    pub config_handler: Option<cosmic_config::Config>,
    pub config: Config,
}
{{- end}}

impl cosmic::Application for Window {
    type Executor = cosmic::SingleThreadExecutor;
    type Flags = {{if .Config}} Flags  {{else}} () {{end}} ;
    type Message = Message;
    const APP_ID: &'static str = ID;

    fn core(&self) -> &Core {
        &self.core
    }

    fn core_mut(&mut self) -> &mut Core {
        &mut self.core
    }

    fn init(
        core: Core,
        {{if .Config}}flags{{else}}_flags{{end}}: Self::Flags,
    ) -> (Self, Command<cosmic::app::Message<Self::Message>>) {
        let window = Window {
            core,
            {{if .Config -}}
            config: flags.config,
            config_handler: flags.config_handler,
            {{- end}}
            popup: None,
            {{if .Example -}} example_row: false, {{- end}}
        };
        (window, Command::none())
    }

    fn on_close_requested(&self, id: window::Id) -> Option<Message> {
        Some(Message::PopupClosed(id))
    }

    fn update(&mut self, message: Self::Message) -> Command<cosmic::app::Message<Self::Message>> {
        {{if .Config}}
        // Helper for updating config values efficiently
        #[allow(unused_macros)]
        macro_rules! config_set {
            ($name: ident, $value: expr) => {
                match &self.config_handler {
                    Some(config_handler) => {
                        match paste::paste! { self.config.[<set_ $name>](config_handler, $value) } {
                            Ok(_) => {}
                            Err(err) => {
                                eprintln!("failed to save config {:?}: {}", stringify!($name), err);
                            }
                        }
                    }
                    None => {
                        self.config.$name = $value;
                        eprintln!(
                            "failed to save config {:?}: no config handler",
                            stringify!($name),
                        );
                    }
                }
            };
        }
        {{end}}
        match message {
            {{if .Config}}
            Message::Config(config) => {
                if config != self.config {
                    self.config = config
                }
            }
            {{end}}
            Message::TogglePopup => {
                return if let Some(p) = self.popup.take() {
                    destroy_popup(p)
                } else {
                    let new_id = Id::unique();
                    self.popup.replace(new_id);
                    let mut popup_settings =
                        self.core
                            .applet
                            .get_popup_settings(Id::MAIN, new_id, None, None, None);
                    popup_settings.positioner.size_limits = Limits::NONE
                        .max_width(372.0)
                        .min_width(300.0)
                        .min_height(200.0)
                        .max_height(1080.0);
                    get_popup(popup_settings)
                }
            }
            Message::PopupClosed(id) => {
                if self.popup.as_ref() == Some(&id) {
                    self.popup = None;
                }
            }
            {{if .Example -}}
            Message::ToggleExampleRow(toggled) => self.example_row = toggled,
            {{- end}}
        }
        Command::none()
    }

    fn view(&self) -> Element<Self::Message> {
        self.core
            .applet
            .icon_button("{{.IconName}}")
            .on_press(Message::TogglePopup)
            .into()
    }

    fn view_window(&self, _id: Id) -> Element<Self::Message> {
        let content_list = widget::list_column()
            .padding(5)
            .spacing(0)
            {{if .Example -}}
            .add(widget::settings::item(
                "Example row",
                widget::toggler(None, self.example_row, |value| {
                    Message::ToggleExampleRow(value)
                }),
            )){{- end}};

        self.core.applet.popup_container(content_list).into()
    }
    fn subscription(&self) -> Subscription<Self::Message> {
        {{if .Config}}
        struct ConfigSubscription;
        return cosmic_config::config_subscription(
            std::any::TypeId::of::<ConfigSubscription>(),
            Self::APP_ID.into(),
            CONFIG_VERSION,
        )
        .map(|update| {
            if !update.errors.is_empty() {
                eprintln!(
                    "errors loading config {:?}: {:?}",
                    update.keys, update.errors
                );
            }
            Message::Config(update.config)
        });
        {{else}}
        Subscription::none()
        {{end}}

    }

    fn style(&self) -> Option<<Theme as application::StyleSheet>::Style> {
        Some(cosmic::applet::style())
    }
}
