const React = require('react');
const Spring = require('react-motion').Spring;

const AppActions = require('../actions/app-actions');
const NodeColorMixin = require('../mixins/node-color-mixin');

const Node = React.createClass({
  mixins: [
    NodeColorMixin
  ],

  getTransform: function(coords) {
    return 'translate(' + coords.x + ',' + coords.y + ')';
  },

  render: function() {
    const props = this.props;
    const coords = {x: props.dx, y: props.dy};
    const getTransform = this.getTransform;
    const scale = props.scale;
    const textOffsetX = 0;
    const textOffsetY = scale(0.5) + 18;
    const isPseudo = !!props.pseudo;
    const color = isPseudo ? '' : this.getNodeColor(props.label);
    const onClick = isPseudo ? null : props.onClick;
    const onMouseEnter = isPseudo ? null : this.handleMouseEnter;
    const onMouseLeave = isPseudo ? null : this.handleMouseLeave;
    const classNames = ['node'];

    if (props.highlighted) {
      classNames.push('highlighted');
    }

    if (props.pseudo) {
      classNames.push('pseudo');
    }

    return (
      <Spring endValue={coords}>
        {function(interpolated) {
          return (
            <g className={classNames.join(' ')} transform={getTransform(interpolated)} id={props.id}
              onClick={onClick} onMouseEnter={onMouseEnter} onMouseLeave={onMouseLeave}>
              {props.highlighted && <circle r={scale(0.7)} className="highlighted"></circle>}
              <circle r={scale(0.5)} className="border" stroke={color}></circle>
              <circle r={scale(0.45)} className="shadow"></circle>
              <circle r={Math.max(2, scale(0.125))} className="node"></circle>
              <text className="node-label" textAnchor="middle" x={textOffsetX} y={textOffsetY}>{props.label}</text>
              <text className="node-sublabel" textAnchor="middle" x={textOffsetX} y={textOffsetY + 17}>{props.subLabel}</text>
            </g>
          );
        }}
      </Spring>
    );
  },

  handleMouseEnter: function(ev) {
    AppActions.enterNode(ev.currentTarget.id);
  },

  handleMouseLeave: function(ev) {
    AppActions.leaveNode(ev.currentTarget.id);
  }

});

module.exports = Node;
