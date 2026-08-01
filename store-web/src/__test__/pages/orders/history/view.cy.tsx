import HistoryView from '@/app/orders/history/view'

describe('<HistoryView />', () => {
  it('Should show the loading state while the request is in flight', () => {
    cy.intercept('GET', '**/api/v1/order/history', {
      delay: 500,
      statusCode: 200,
      body: []
    }).as('getOrderHistory')

    cy.mount(<HistoryView />)

    cy.get('#order-history-loading')
      .should('be.visible')
      .should('have.text', 'Loading your orders...')

    cy.wait('@getOrderHistory')
  })

  it('Should show "You have no orders yet" when there are no orders', () => {
    cy.intercept('GET', '**/api/v1/order/history', {
      statusCode: 200,
      body: []
    }).as('getOrderHistory')

    cy.mount(<HistoryView />)

    cy.wait('@getOrderHistory')
    cy.get('#order-history-empty')
      .should('be.visible')
      .should('have.text', 'You have no orders yet')
    cy.get('#order-history-table').should('not.exist')
  })

  it('Should render the orders returned by the API', () => {
    cy.intercept('GET', '**/api/v1/order/history', {
      statusCode: 200,
      body: [
        {
          order_number: 2601069522002002,
          status: 'paid',
          sub_total_price: 5246.22,
          total_price: 5256.22,
          burn_point: 0,
          earn_point: 52,
          tracking_no: 'KR-304590466',
          updated: '2026-02-28T18:58:44Z'
        },
        {
          order_number: 2601069522001001,
          status: 'created',
          sub_total_price: 4314.6,
          total_price: 4364.6,
          burn_point: 10,
          earn_point: 0,
          tracking_no: '',
          updated: '2026-02-14T01:40:32Z'
        }
      ]
    }).as('getOrderHistory')

    cy.mount(<HistoryView />)

    cy.wait('@getOrderHistory')
    cy.get('#order-history-empty').should('not.exist')
    cy.get('#order-history-table').should('be.visible')

    cy.get('#order-history-row-2601069522002002').within(() => {
      cy.contains('td', '2601069522002002')
      cy.contains('td', 'paid')
      cy.contains('td', '5,256.22')
      cy.contains('td', '52')
      cy.contains('td', 'KR-304590466')
    })

    cy.get('#order-history-row-2601069522001001').within(() => {
      cy.contains('td', 'created')
      cy.contains('td', '4,364.6')
      cy.contains('td', '10')
      cy.contains('td', '-')
    })
  })

  it('Should only show the Confirm Receipt button for "paid" rows, and a static label for "completed" rows', () => {
    cy.intercept('GET', '**/api/v1/order/history', {
      statusCode: 200,
      body: [
        {
          order_number: 1001,
          status: 'paid',
          sub_total_price: 100,
          total_price: 100,
          burn_point: 0,
          earn_point: 2,
          tracking_no: 'KR-1',
          updated: '2026-02-28T18:58:44Z'
        },
        {
          order_number: 1002,
          status: 'completed',
          sub_total_price: 100,
          total_price: 100,
          burn_point: 0,
          earn_point: 2,
          tracking_no: 'KR-2',
          updated: '2026-02-28T18:58:44Z'
        },
        {
          order_number: 1003,
          status: 'created',
          sub_total_price: 100,
          total_price: 100,
          burn_point: 0,
          earn_point: 0,
          tracking_no: '',
          updated: '2026-02-28T18:58:44Z'
        },
        {
          order_number: 1004,
          status: 'cancel',
          sub_total_price: 100,
          total_price: 100,
          burn_point: 0,
          earn_point: 0,
          tracking_no: '',
          updated: '2026-02-28T18:58:44Z'
        }
      ]
    }).as('getOrderHistory')

    cy.mount(<HistoryView />)
    cy.wait('@getOrderHistory')

    cy.get('#order-history-confirm-1001')
      .should('be.visible')
      .should('not.be.disabled')
      .should('have.text', 'Confirm Receipt')

    cy.get('#order-history-confirm-1002').should('not.exist')
    cy.get('#order-history-confirmed-1002')
      .should('be.visible')
      .should('have.text', 'Confirmed')

    cy.get('#order-history-confirm-1003').should('not.exist')
    cy.get('#order-history-confirmed-1003').should('not.exist')

    cy.get('#order-history-confirm-1004').should('not.exist')
    cy.get('#order-history-confirmed-1004').should('not.exist')
  })

  it('Should confirm receipt and flip the row to "Confirmed" on success', () => {
    cy.intercept('GET', '**/api/v1/order/history', {
      statusCode: 200,
      body: [
        {
          order_number: 2002,
          status: 'paid',
          sub_total_price: 100,
          total_price: 100,
          burn_point: 0,
          earn_point: 2,
          tracking_no: 'KR-2002',
          updated: '2026-02-28T18:58:44Z'
        }
      ]
    }).as('getOrderHistory')

    cy.intercept('POST', '**/api/v1/order/2002/confirmReceipt', {
      delay: 300,
      statusCode: 200,
      body: { order_number: 2002, status: 'completed' }
    }).as('confirmReceipt')

    cy.mount(<HistoryView />)
    cy.wait('@getOrderHistory')

    cy.get('#order-history-confirm-2002').click()

    // disabled while the request is in flight
    cy.get('#order-history-confirm-2002').should('be.disabled')

    cy.wait('@confirmReceipt')

    // row flips to Confirmed, without a full refetch of the list
    cy.get('#order-history-confirm-2002').should('not.exist')
    cy.get('#order-history-confirmed-2002')
      .should('be.visible')
      .should('have.text', 'Confirmed')
    cy.get('#order-history-row-2002').within(() => {
      cy.contains('td', 'completed')
    })
    cy.get('@getOrderHistory.all').should('have.length', 1)
  })

  it('Should show an inline error and keep the button clickable if confirmation fails', () => {
    cy.intercept('GET', '**/api/v1/order/history', {
      statusCode: 200,
      body: [
        {
          order_number: 3003,
          status: 'paid',
          sub_total_price: 100,
          total_price: 100,
          burn_point: 0,
          earn_point: 2,
          tracking_no: 'KR-3003',
          updated: '2026-02-28T18:58:44Z'
        }
      ]
    }).as('getOrderHistory')

    cy.intercept('POST', '**/api/v1/order/3003/confirmReceipt', {
      statusCode: 500,
      body: { error: 'Internal Server Error' }
    }).as('confirmReceipt')

    cy.mount(<HistoryView />)
    cy.wait('@getOrderHistory')

    cy.get('#order-history-confirm-3003').click()
    cy.wait('@confirmReceipt')

    cy.get('#order-history-confirm-error-3003').should('be.visible')
    cy.get('#order-history-confirm-3003')
      .should('be.visible')
      .should('not.be.disabled')
      .should('have.text', 'Confirm Receipt')
    cy.get('#order-history-row-3003').within(() => {
      cy.contains('td', 'paid')
    })
  })

  it('Should show the empty state if the API call fails', () => {
    cy.intercept('GET', '**/api/v1/order/history', {
      statusCode: 500,
      body: { error: 'Internal Server Error' }
    }).as('getOrderHistory')

    cy.mount(<HistoryView />)

    cy.wait('@getOrderHistory')
    cy.get('#order-history-empty').should('be.visible')
  })
})
