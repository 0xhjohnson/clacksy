const mobileMenuBtn = document.getElementById('mobile-menu-btn')
const mobileMenu = makeMobileMenu(mobileMenuBtn)
mobileMenuBtn.addEventListener('click', mobileMenu.handleClick)

const userMenuButton = document.getElementById('user-menu-button')
// Only authenticated (signed in) users have user menu button displayed
if (userMenuButton) {
  const userMenu = makeUserMenu(userMenuButton)
  userMenuButton.addEventListener('click', userMenu.handleClick)
}

function makeMobileMenu(mobileMenuBtn) {
  let isMobileMenuOpen = false

  return {
    get isOpen() {
      return isMobileMenuOpen
    },
    handleClick: function () {
      const mobileMenuEl = document.getElementById('mobile-menu')
      const srEl = document.getElementById('mobile-menu-sr-text')
      const hamburgerIconEl = document.getElementById('hamburger-icon')
      const mobileCloseIconEl = document.getElementById('mobile-close-icon')

      isMobileMenuOpen = !isMobileMenuOpen

      if (isMobileMenuOpen) {
        hideElement(hamburgerIconEl)
        showElement(mobileCloseIconEl)

        mobileMenuBtn.setAttribute('aria-expanded', 'true')
        showElement(mobileMenuEl)
      } else {
        hideElement(mobileCloseIconEl)
        showElement(hamburgerIconEl)

        mobileMenuBtn.removeAttribute('aria-expanded')
        hideElement(mobileMenuEl)
      }

      srEl.innerText = `${this.isOpen ? 'Close' : 'Open'} main menu`
    }
  }
}

function makeUserMenu(userMenuEl) {
  let isProfileOpen = false

  return {
    get isOpen() {
      return isProfileOpen
    },
    handleClick: function () {
      const menuListEl = document.getElementById('user-menu-list')
      const srEl = document.getElementById('user-menu-sr-text')

      isProfileOpen = !isProfileOpen

      if (isProfileOpen) {
        userMenuEl.setAttribute('aria-expanded', 'true')
        showElement(menuListEl)
      } else {
        userMenuEl.removeAttribute('aria-expanded')
        hideElement(menuListEl)
      }

      srEl.innerText = `${this.isOpen ? 'Close' : 'Open'} user menu`
    }
  }
}

function showElement(el) {
  el.classList.remove('hidden')
}

function hideElement(el) {
  el.classList.remove('block')
  el.classList.add('hidden')
}
