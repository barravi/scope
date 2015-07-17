const React = require('react');
const tweenState = require('react-tween-state');

const AppActions = require('../actions/app-actions');

const Edge = React.createClass({
  mixins: [
    tweenState.Mixin
  ],

  getInitialState: function() {
    return {
      x1: 0,
      x2: 0,
      y1: 0,
      y2: 0
    };
  },

  componentWillMount: function() {
    // initial node position when rendered the first time
    this.setState({
      x1: this.props.points[0].x,
      x2: this.props.points[1].x,
      y1: this.props.points[0].y,
      y2: this.props.points[1].y
    });
  },

  componentWillReceiveProps: function(nextProps) {
    // animate node transition to next position
    this.tweenState('x1', {
      easing: tweenState.easingTypes.easeInOutQuad,
      duration: 500,
      endValue: nextProps.points[0].x
    });
    this.tweenState('x2', {
      easing: tweenState.easingTypes.easeInOutQuad,
      duration: 500,
      endValue: nextProps.points[1].x
    });
    this.tweenState('y1', {
      easing: tweenState.easingTypes.easeInOutQuad,
      duration: 500,
      endValue: nextProps.points[0].y
    });
    this.tweenState('y2', {
      easing: tweenState.easingTypes.easeInOutQuad,
      duration: 500,
      endValue: nextProps.points[1].y
    });
  },

  render: function() {
    const className = this.props.highlighted ? 'edge highlighted' : 'edge';
    const path = 'M' +
      this.getTweeningValue('x1') + ',' +
      this.getTweeningValue('y1') + 'L' +
      this.getTweeningValue('x2') + ',' +
      this.getTweeningValue('y2');

    return (
      <g className={className} onMouseEnter={this.handleMouseEnter} onMouseLeave={this.handleMouseLeave} id={this.props.id}>
        <path d={path} className="shadow" />
        <path d={path} className="link" />
      </g>
    );
  },

  handleMouseEnter: function(ev) {
    AppActions.enterEdge(ev.currentTarget.id);
  },

  handleMouseLeave: function(ev) {
    AppActions.leaveEdge(ev.currentTarget.id);
  }

});

module.exports = Edge;
