  @doc """
  Example:
  <.navbar drawer_toggle_id="app-drawer">
    <:start><a class="btn btn-ghost text-xl">Brand</a></:start>
    <:center><span>Center</span></:center>
    <:right><button class="btn">Action</button></:right>
  </.navbar>
  """
  attr :class, :any, default: nil
  attr :drawer_toggle_id, :string, default: nil
  slot :start
  slot :center
  slot :right
  def navbar(assigns) do
    ~H"""
    <div class={["navbar bg-base-200", @class]}>
      <div class="navbar-start">
        <label
          :if={@drawer_toggle_id}
          for={@drawer_toggle_id}
          class="btn btn-ghost btn-square lg:hidden"
          aria-label={gettext("open sidebar")}
        >
          <.icon name="hero-bars-3" class="size-6" />
        </label>
        {render_slot(@start)}
      </div>
      <div class="navbar-center">{render_slot(@center)}</div>
      <div class="navbar-end">{render_slot(@right)}</div>
    </div>
    """
  end

  @doc """
  Example:
  <.drawer id="app-drawer" class="lg:drawer-open">
    <:content>
      <.navbar drawer_toggle_id="app-drawer" />
      <div class="p-4">Page content</div>
    </:content>
    <:sidebar>
      <ul class="menu p-4 w-80 min-h-full bg-base-100">
        <li><.link navigate={~p"/"}>Dashboard</.link></li>
        <li><.link navigate={~p"/settings"}>Settings</.link></li>
      </ul>
    </:sidebar>
  </.drawer>
  """
  attr :id, :string, required: true
  attr :right, :boolean, default: false
  attr :open, :boolean, default: false
  attr :class, :any, default: nil
  attr :content_class, :any, default: nil
  attr :sidebar_class, :any, default: "bg-base-100 text-base-content w-80 min-h-full p-4"
  slot :content, required: true
  slot :sidebar, required: true
  def drawer(assigns) do
    ~H"""
    <div class={[
          "drawer",
          @right && "drawer-end",
          @open && "drawer-open",
          @class
        ]}>
      <input id={@id} type="checkbox" class="drawer-toggle" />
      <div class={["drawer-content", @content_class]}>
        {render_slot(@content)}
      </div>
      <div class="drawer-side">
        <label for={@id} class="drawer-overlay" aria-label={gettext("close sidebar")} />
        <div class={@sidebar_class}>
          {render_slot(@sidebar)}
        </div>
      </div>
    </div>
    """
  end

  @doc """
  Example:
  <.dropdown placement="end">
    <:trigger><button class="btn">Actions</button></:trigger>
    <:content>
      <ul class="menu menu-sm gap-1">
        <li><a>Edit</a></li>
        <li><a>Delete</a></li>
      </ul>
    </:content>
  </.dropdown>
  """
  attr :class, :any, default: nil
  attr :content_class, :any, default: "dropdown-content bg-base-300 rounded-box mt-2 w-56 p-2 shadow"
  attr :placement, :string, default: "end", values: ~w(start center end top bottom left right)
  attr :hover, :boolean, default: false
  attr :open, :boolean, default: false
  slot :trigger, required: true
  slot :content, required: true
  def dropdown(assigns) do
    ~H"""
    <div class={[
          "dropdown",
          @hover && "dropdown-hover",
          @placement in ["start","center","end"] && "dropdown-#{@placement}",
          @placement in ["top","bottom","left","right"] && "dropdown-#{@placement}",
          @open && "dropdown-open",
          @class
        ]}>
      <div tabindex="0" role="button">
        {render_slot(@trigger)}
      </div>
      <div tabindex="0" class={@content_class}>
        {render_slot(@content)}
      </div>
    </div>
    """
  end

  @doc """
  Example:
  <.tabs id="profile-tabs" style="box">
    <:tab label="Profile" icon="hero-user" active><p>Profile content</p></:tab>
    <:tab label="Security"><p>Security content</p></:tab>
  </.tabs>
  """
  attr :id, :string, required: true
  attr :class, :any, default: nil
  attr :style, :string, default: nil, values: [nil, "box", "border", "lift"]
  attr :placement, :string, default: "top", values: ~w(top bottom)
  slot :tab, required: true do
    attr :label, :string, required: true
    attr :icon, :string
    attr :active, :boolean
    attr :disabled, :boolean
  end
  def tabs(assigns) do
    ~H"""
    <div role="tablist" class={[
          "tabs",
          @style == "box" && "tabs-box",
          @style == "border" && "tabs-border",
          @style == "lift" && "tabs-lift",
          @placement == "top" && "tabs-top",
          @placement == "bottom" && "tabs-bottom",
          @class
        ]}>
      <%= for {tab, idx} <- Enum.with_index(@tab) do %>
        <input
          type="radio"
          name={@id}
          role="tab"
          class={["tab", tab[:disabled] && "tab-disabled"]}
          aria-label={tab[:label]}
          checked={tab[:active] || idx == 0}
          disabled={tab[:disabled]}
        />
        <div role="tabpanel" class="tab-content">
          <div class="flex items-center gap-2 mb-2">
            <.icon :if={tab[:icon]} name={tab[:icon]} class="size-4" />
            <span class="font-semibold">{tab[:label]}</span>
          </div>
          {render_slot(tab)}
        </div>
      <% end %>
    </div>
    """
  end

  @doc """
  Example:
  <div class="flex items-center gap-3">
    <.loading variant="spinner" />
    <.loading variant="dots" size="sm" />
    <.loading variant="ring" size="lg" />
  </div>
  """
  attr :variant, :string, default: "spinner", values: ~w(spinner dots ring ball bars infinity)
  attr :size, :string, default: "md", values: ~w(xs sm md lg xl)
  attr :label, :string, default: nil
  attr :class, :any, default: nil
  def loading(assigns) do
    ~H"""
    <span
      class={[
        "loading",
        "loading-#{@variant}",
        "loading-#{@size}",
        @class
      ]}
      aria-live="polite"
      aria-busy="true"
    />
    <span :if={@label} class="sr-only">{@label}</span>
    """
  end

  @doc """
  Example:
  <.progress value={42} color="primary" class="w-56" />
  """
  attr :value, :integer, required: true
  attr :max, :integer, default: 100
  attr :color, :string, default: nil,
    values: [nil, "neutral", "primary", "secondary", "accent", "info", "success", "warning", "error"]
  attr :class, :any, default: nil
  def progress(assigns) do
    ~H"""
    <progress
      class={[
        "progress",
        @color && "progress-#{@color}",
        @class
      ]}
      value={@value}
      max={@max}
      aria-valuenow={@value}
      aria-valuemin="0"
      aria-valuemax={@max}
    >
      {@value}%
    </progress>
    """
  end

  @doc """
  Example:
  <.radial_progress value={68} style="--size:6rem; --thickness:8px;" />
  """
  attr :value, :integer, required: true
  attr :label, :string, default: nil
  attr :style, :string, default: nil
  attr :class, :any, default: nil
  def radial_progress(assigns) do
    ~H"""
    <div
      class={["radial-progress", @class]}
      style={[
        "--value: #{@value};",
        @style
      ] |> Enum.reject(&is_nil/1) |> Enum.join(" ")}
      role="progressbar"
      aria-valuenow={@value}
    >
      {if @label, do: @label, else: "#{@value}%"}
    </div>
    """
  end

  @doc """
  Example:
  <.menu class="w-56">
    <:item navigate={~p"/"} icon="hero-home" active>Home</:item>
    <:item navigate={~p"/clients"} icon="hero-users">Clients</:item>
  </.menu>
  """
  attr :class, :any, default: nil
  attr :direction, :string, default: "vertical", values: ~w(vertical horizontal)
  slot :item, required: true do
    attr :href, :string
    attr :navigate, :any
    attr :patch, :any
    attr :active, :boolean
    attr :disabled, :boolean
    attr :icon, :string
  end
  def menu(assigns) do
    ~H"""
    <ul class={[
          "menu",
          @direction == "horizontal" && "menu-horizontal",
          @class
        ]}>
      <li :for={item <- @item}>
        <.link
          :if={item[:href] || item[:navigate] || item[:patch]}
          href={item[:href]}
          navigate={item[:navigate]}
          patch={item[:patch]}
          class={[
            item[:active] && "menu-active",
            item[:disabled] && "menu-disabled"
          ]}
        >
          <.icon :if={item[:icon]} name={item[:icon]} class="size-4" />
          {render_slot(item)}
        </.link>
        <button
          :if={! (item[:href] || item[:navigate] || item[:patch])}
          class={[
            item[:active] && "menu-active",
            item[:disabled] && "menu-disabled"
          ]}
        >
          <.icon :if={item[:icon]} name={item[:icon]} class="size-4" />
          {render_slot(item)}
        </button>
      </li>
    </ul>
    """
  end

  @doc """
  Example:
  <.breadcrumbs>
    <:crumb navigate={~p"/"}>Home</:crumb>
    <:crumb>Dashboard</:crumb>
  </.breadcrumbs>
  """
  attr :class, :any, default: nil
  slot :crumb, required: true do
    attr :href, :string
    attr :navigate, :any
    attr :patch, :any
  end
  def breadcrumbs(assigns) do
    ~H"""
    <div class={["breadcrumbs", @class]}>
      <ul>
        <li :for={c <- @crumb}>
          <.link
            :if={c[:href] || c[:navigate] || c[:patch]}
            href={c[:href]}
            navigate={c[:navigate]}
            patch={c[:patch]}
            class="link"
          >
            {render_slot(c)}
          </.link>
          <span :if={! (c[:href] || c[:navigate] || c[:patch])} class="text-base-content/70">
            {render_slot(c)}
          </span>
        </li>
      </ul>
    </div>
    """
  end

  @doc """
  Example:
  <.badge color="success" style="soft">Active</.badge>
  """
  attr :style, :string, default: nil, values: [nil, "outline", "dash", "soft", "ghost"]
  attr :color, :string, default: nil,
    values: [nil, "neutral", "primary", "secondary", "accent", "info", "success", "warning", "error"]
  attr :size, :string, default: nil, values: [nil, "xs", "sm", "md", "lg", "xl"]
  attr :class, :any, default: nil
  slot :inner_block, required: true
  def badge(assigns) do
    ~H"""
    <span class={[
          "badge",
          @style && "badge-#{@style}",
          @color && "badge-#{@color}",
          @size && "badge-#{@size}",
          @class
        ]}>
      {render_slot(@inner_block)}
    </span>
    """
  end

  @doc """
  Example:
  <.avatar src="https://picsum.photos/80" alt="U" class="w-10 h-10" />
  """
  attr :src, :string, default: nil
  attr :alt, :string, default: ""
  attr :placeholder, :boolean, default: false
  attr :class, :any, default: nil
  attr :rounded, :boolean, default: true
  slot :placeholder_content
  def avatar(assigns) do
    ~H"""
    <div class={[
          "avatar",
          @placeholder && "avatar-placeholder",
          @class
        ]}>
      <div class={if @rounded, do: "rounded-full", else: "rounded"}>
        <%= if @placeholder do %>
          <div class="flex items-center justify-center w-full h-full ring-primary rounded-full ring-2">{render_slot(@placeholder_content)}</div>
        <% else %>
          <img src={@src} alt={@alt} />
        <% end %>
      </div>
    </div>
    """
  end

  @doc """
  Example:
  <.pagination>
    <:item patch={~p"/projects?page=1"}>&laquo;</:item>
    <:item patch={~p"/projects?page=1"} active>1</:item>
    <:item patch={~p"/projects?page=2"}>2</:item>
    <:item patch={~p"/projects?page=2"}>&raquo;</:item>
  </.pagination>
  """
  attr :class, :any, default: nil
  slot :item, required: true do
    attr :href, :string
    attr :navigate, :any
    attr :patch, :any
    attr :active, :boolean
    attr :disabled, :boolean
  end
  def pagination(assigns) do
    ~H"""
    <div class={["join", @class]}>
      <%= for item <- @item do %>
        <.link
          :if={item[:href] || item[:navigate] || item[:patch]}
          href={item[:href]}
          navigate={item[:navigate]}
          patch={item[:patch]}
          class={[
            "join-item btn",
            item[:active] && "btn-active",
            item[:disabled] && "btn-disabled"
          ]}
        >
          {render_slot(item)}
        </.link>
        <button
          :if={! (item[:href] || item[:navigate] || item[:patch])}
          class={[
            "join-item btn",
            item[:active] && "btn-active",
            item[:disabled] && "btn-disabled"
          ]}
        >
          {render_slot(item)}
        </button>
      <% end %>
    </div>
    """
  end

  @doc """
  Example:
  <.card image_src="https://picsum.photos/400/200">
    <:header>Card title</:header>
    <p>Card content</p>
    <:actions><button class="btn btn-primary">Action</button></:actions>
  </.card>
  """
  attr :class, :any, default: nil
  attr :image_src, :string, default: nil
  attr :image_alt, :string, default: ""
  attr :side, :boolean, default: false
  slot :header
  slot :inner_block, required: true
  slot :actions
  def card(assigns) do
    ~H"""
    <div class={[
          "card bg-base-200 shadow-sm",
          @side && "card-side",
          !@side && "w-96",
          @class
        ]}>
      <figure :if={@image_src}><img src={@image_src} alt={@image_alt} /></figure>
      <div class="card-body">
        <h2 :if={@header != []} class="card-title">{render_slot(@header)}</h2>
        {render_slot(@inner_block)}
        <div :if={@actions != []} class="card-actions">{render_slot(@actions)}</div>
      </div>
    </div>
    """
  end

  @doc """
  Example:
  <.dock>
    <:item active>
      <.icon name="hero-home" class="size-6" />
      <span class="dock-label">Home</span>
    </:item>
    <:item>
      <.icon name="hero-bell" class="size-6" />
      <span class="dock-label">Alerts</span>
    </:item>
  </.dock>
  """
  attr :class, :any, default: nil
  attr :size, :string, default: nil, values: [nil, "xs", "sm", "md", "lg", "xl"]
  slot :item, required: true do
    attr :active, :boolean
  end
  def dock(assigns) do
    ~H"""
    <div class={[
          "dock",
          @size && "dock-#{@size}",
          @class
        ]}>
      <button :for={item <- @item} class={[item[:active] && "dock-active"]}>
        {render_slot(item)}
      </button>
    </div>
    """
  end

  @doc """
  Example:
  <.collapse title="More info" arrow open>
    <p>Hidden content</p>
  </.collapse>
  """
  attr :title, :string, default: nil
  attr :open, :boolean, default: false
  attr :arrow, :boolean, default: false
  attr :plus, :boolean, default: false
  attr :class, :any, default: nil
  slot :inner_block, required: true
  def collapse(assigns) do
    ~H"""
    <details class={[
              "collapse",
              @arrow && "collapse-arrow",
              @plus && "collapse-plus",
              @class
            ]} open={@open}>
      <summary class="collapse-title">{@title}</summary>
      <div class="collapse-content">
        {render_slot(@inner_block)}
      </div>
    </details>
    """
  end

  @doc """
  Example:
  <.tooltip text="Hello" placement="right">
    <button class="btn">Hover</button>
  </.tooltip>
  """
  attr :text, :string, required: true
  attr :placement, :string, default: nil, values: [nil, "top", "bottom", "left", "right"]
  attr :open, :boolean, default: false
  attr :class, :any, default: nil
  slot :inner_block, required: true
  def tooltip(assigns) do
    ~H"""
    <div
      class={[
        "tooltip",
        @placement && "tooltip-#{@placement}",
        @open && "tooltip-open",
        @class
      ]}
      data-tip={@text}
    >
      {render_slot(@inner_block)}
    </div>
    """
  end

  @doc """
  Example:
  <.toggle id="news" name="news" color="primary" checked />
  """
  attr :id, :string, default: nil
  attr :name, :string, default: nil
  attr :checked, :boolean, default: false
  attr :disabled, :boolean, default: false
  attr :color, :string, default: nil,
    values: [nil, "neutral", "primary", "secondary", "accent", "info", "success", "warning", "error"]
  attr :size, :string, default: nil, values: [nil, "xs", "sm", "md", "lg", "xl"]
  attr :class, :any, default: nil
  attr :rest, :global, include: ~w(value)
  def toggle(assigns) do
    ~H"""
    <input
      type="checkbox"
      id={@id}
      name={@name}
      class={[
        "toggle",
        @color && "toggle-#{@color}",
        @size && "toggle-#{@size}",
        @class
      ]}
      checked={@checked}
      disabled={@disabled}
      {@rest}
    />
    """
  end

  @doc """
  Example:
  <.chat placement="start" color="primary">
    <:image><.avatar src="https://picsum.photos/48" class="w-10 h-10" /></:image>
    <:header>Alice</:header>
    Hello there
    <:footer>2 min ago</:footer>
  </.chat>
  """
  attr :placement, :string, default: "start", values: ~w(start end)
  attr :color, :string, default: nil,
    values: [nil, "neutral", "primary", "secondary", "accent", "info", "success", "warning", "error"]
  slot :image
  slot :header
  slot :footer
  slot :inner_block, required: true
  def chat(assigns) do
    ~H"""
    <div class={["chat", "chat-#{@placement}"]}>
      <div :if={@image != []} class="chat-image avatar">
        {render_slot(@image)}
      </div>
      <div :if={@header != []} class="chat-header">{render_slot(@header)}</div>
      <div class={["chat-bubble", @color && "chat-bubble-#{@color}"]}>
        {render_slot(@inner_block)}
      </div>
      <div :if={@footer != []} class="chat-footer">{render_slot(@footer)}</div>
    </div>
    """
  end

  def show_modal(id) when is_binary(id), do: JS.dispatch("show-dialog-modal", to: "##{id}")
  def show_modal(%JS{} = js, id), do: JS.dispatch(js, "show-dialog-modal", to: "##{id}")
  def close_modal(id) when is_binary(id), do: JS.dispatch("hide-dialog-modal", to: "##{id}")
  def close_modal(%JS{} = js, id), do: JS.dispatch(js, "hide-dialog-modal", to: "##{id}")

  def cancel_modal(id) when is_binary(id),
    do: JS.exec("data-cancel", to: "##{id}") |> close_modal(id)

  def cancel_modal(%JS{} = js, id),
    do: JS.exec(js, "data-cancel", to: "##{id}") |> close_modal(id)

  @doc """
  Modal con <dialog> + DaisyUI.

  - id:       id del dialog
  - on_cancel: JS da eseguire su annullo/ESC/click-away (es. reset form, push, ecc.)
  - slot:     riceve
  """
  attr :id, :string, required: true
  attr :on_cancel, JS, default: %JS{}
  attr :class, :any, default: nil
  slot :inner_block, required: true

  def modal(assigns) do
    ~H"""
    <dialog
      id={@id}
      class={["modal", @class]}
      phx-hook="Dialog"
      phx-mounted={JS.ignore_attributes(["open"])}
      phx-remove={
        JS.remove_attribute("open")
        |> JS.transition({"ease-out duration-200", "opacity-100", "opacity-0"}, time: 0)
      }
      data-cancel={@on_cancel}
      phx-window-keydown={cancel_modal(@id)}
      phx-key="escape"
    >
      <.focus_wrap
        id={"#{@id}-box"}
        class="modal-box"
        phx-click-away={cancel_modal(@id)}
      >
        <button
          type="button"
          class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2"
          phx-click={cancel_modal(@id)}
          aria-label="close"
        >
          <.icon name="hero-x-mark" class="w-5 h-5" />
        </button>

        {render_slot(@inner_block, %{close: close_modal(@id), cancel: cancel_modal(@id)})}
      </.focus_wrap>
    </dialog>
    """
  end
