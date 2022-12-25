-- This Source Code Form is subject to the terms of the Mozilla Public
-- License, v. 2.0. If a copy of the MPL was not distributed with this
-- file, You can obtain one at https://mozilla.org/MPL/2.0/.

function Link(el)
  el.target = el.target:gsub("%.md$", ".html")
  return el
end

function Header(el)
  name = pandoc.utils.stringify(el.content)
  if name:find("(chipset vendor)") or name:find("(manufacturer)") or name:find("(device)") then
    el.classes:insert("device")
  end
  return el
end
